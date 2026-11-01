package com.ridehailing.common.events;

import com.ridehailing.common.jdbc.Timestamps;

import java.time.Clock;
import java.time.Instant;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.jdbc.core.simple.JdbcClient;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;

/**
 * Deletes what delivery no longer needs: outbox rows published longer ago than
 * {@code events.retention.published-outbox} (their payloads hold pickup and driver locations, GDPR), and de-duplication
 * markers older than {@code events.retention.processed-events}. Unpublished rows are never deleted. Each batch is its
 * own statement and skips rows another transaction holds, so a backlog never blocks the relay or consumers, and
 * instances can run it concurrently.
 */
@Component
@RequiredArgsConstructor
@Slf4j
public class EventRetention {
    private final JdbcClient jdbc;
    private final Clock clock;
    private final EventRetentionProperties properties;

    @Scheduled(fixedDelayString = "${events.retention.interval-ms:3600000}")
    public void prune() {
        try {
            Instant now = clock.instant();
            int outboxRows = deleteInBatches("""
                    DELETE FROM outbox
                    WHERE id IN (SELECT id
                                 FROM outbox
                                 WHERE published_at < :cutoff
                                 ORDER BY id
                                 LIMIT :limit
                                 FOR UPDATE SKIP LOCKED)
                    """, now.minus(properties.publishedOutbox()));
            int markers = deleteInBatches("""
                    DELETE FROM processed_events
                    WHERE (consumer, event_id) IN (SELECT consumer, event_id
                                                   FROM processed_events
                                                   WHERE processed_at < :cutoff
                                                   LIMIT :limit
                                                   FOR UPDATE SKIP LOCKED)
                    """, now.minus(properties.processedEvents()));
            if (outboxRows + markers > 0) {
                log.info("Event retention deleted {} published outbox rows and {} de-duplication markers",
                        outboxRows, markers);
            }
        } catch (RuntimeException exception) {
            log.warn("Event retention failed; it runs again in the next cycle: {}", exception.toString());
        }
    }

    private int deleteInBatches(String sql, Instant cutoff) {
        int total = 0;
        int deleted;
        do {
            deleted = jdbc.sql(sql)
                    .param("cutoff", Timestamps.toDb(cutoff))
                    .param("limit", properties.batchSize())
                    .update();
            total += deleted;
        } while (deleted == properties.batchSize());
        return total;
    }
}
