package com.ridehailing.common.events;

import com.ridehailing.common.jdbc.Timestamps;
import com.ridehailing.common.logging.LogContext;
import java.time.Clock;
import java.util.ArrayList;
import java.util.List;
import java.util.UUID;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.TimeoutException;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.jdbc.core.simple.JdbcClient;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;
import org.springframework.transaction.support.TransactionTemplate;

/**
 * Publishes committed outbox rows to their topic in insertion order, with their message key, then marks them published.
 * Delivery is at-least-once: a failure before the mark re-sends the whole batch, which consumers de-duplicate by
 * {@code eventId}.
 * <p>
 * One relay publishes at a time across every instance sharing the database: each batch first takes the
 * transaction-scoped advisory lock {@link #LOCK_KEY}, and an instance that cannot take it publishes nothing that round.
 * A batch is claimed only after the previous one committed, which is after Kafka acknowledged it, so events with one key
 * reach Kafka in the order they were recorded. Within a batch every record is sent before waiting once for all
 * acknowledgements; the idempotent producer keeps each partition in send order.
 */
@Component
@RequiredArgsConstructor
@Slf4j
public class OutboxRelay {

    /** The advisory lock every relay on one database takes before publishing ("outbox" in ASCII). */
    public static final long LOCK_KEY = 0x6F7574626F78L;

    private static final long SEND_TIMEOUT_SECONDS = 10;

    private final JdbcClient jdbc;
    private final KafkaTemplate<String, String> kafka;
    private final TransactionTemplate transactions;
    private final Clock clock;
    private final OutboxProperties properties;

    @Scheduled(fixedDelayString = "${outbox.relay.interval-ms:200}")
    public void relayPending() {
        try {
            int published;
            do {
                Integer count = transactions.execute(status -> publishBatch());
                published = count == null ? 0 : count;
            } while (published == properties.batchSize());
        } catch (RuntimeException exception) {
            log.warn("Outbox relay failed; unpublished events will be retried: {}", exception.toString());
        }
    }

    private int publishBatch() {
        if (!Boolean.TRUE.equals(jdbc.sql("SELECT pg_try_advisory_xact_lock(:key)")
                .param("key", LOCK_KEY)
                .query(Boolean.class)
                .single())) {
            log.debug("Another relay holds the outbox lock; nothing published this round");
            return 0;
        }
        List<OutboxRow> rows = jdbc.sql("""
                        SELECT id, event_id, topic, message_key, aggregate_id, event_type, CAST(payload AS text) AS payload
                        FROM outbox
                        WHERE published_at IS NULL
                        ORDER BY id
                        LIMIT :limit
                        FOR UPDATE SKIP LOCKED
                        """)
                .param("limit", properties.batchSize())
                .query((resultSet, rowNumber) -> new OutboxRow(
                        resultSet.getLong("id"),
                        resultSet.getObject("event_id", UUID.class),
                        resultSet.getString("topic"),
                        resultSet.getString("message_key"),
                        resultSet.getString("aggregate_id"),
                        resultSet.getString("event_type"),
                        resultSet.getString("payload")))
                .list();
        if (rows.isEmpty()) {
            return 0;
        }
        send(rows);
        jdbc.sql("UPDATE outbox SET published_at = :now WHERE id IN (:ids)")
                .param("now", Timestamps.toDb(clock.instant()))
                .param("ids", rows.stream().map(OutboxRow::id).toList())
                .update();
        log.info("Published {} outbox events", rows.size());
        return rows.size();
    }

    /** Sends every row, then waits once for all of them to be acknowledged. */
    private void send(List<OutboxRow> rows) {
        List<CompletableFuture<?>> acknowledgements = new ArrayList<>(rows.size());
        for (OutboxRow row : rows) {
            try (LogContext ignored = LogContext.with(LogContext.EVENT_ID, row.eventId())
                    .and("aggregateId", row.aggregateId())
                    .and(LogContext.EVENT_TYPE, row.eventType())) {
                acknowledgements.add(kafka.send(row.topic(), row.messageKey(), row.payload()));
                log.debug("Event {} sent to {}", row.eventType(), row.topic());
            }
        }
        try {
            CompletableFuture.allOf(acknowledgements.toArray(CompletableFuture[]::new))
                    .get(SEND_TIMEOUT_SECONDS, TimeUnit.SECONDS);
        } catch (InterruptedException exception) {
            Thread.currentThread().interrupt();
            throw new IllegalStateException("Interrupted while publishing %d outbox events".formatted(rows.size()),
                    exception);
        } catch (ExecutionException | TimeoutException exception) {
            throw new IllegalStateException("Could not publish a batch of %d outbox events, from id %d"
                    .formatted(rows.size(), rows.getFirst().id()), exception);
        }
    }

    private record OutboxRow(
            long id, UUID eventId, String topic, String messageKey, String aggregateId, String eventType, String payload) {
    }
}
