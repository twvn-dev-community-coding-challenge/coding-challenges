package com.ridehailing.integration;

import static org.assertj.core.api.Assertions.assertThat;

import com.ridehailing.common.events.EventRetention;
import com.ridehailing.common.events.EventRetentionProperties;
import com.ridehailing.common.jdbc.Timestamps;
import com.ridehailing.support.IntegrationTestSupport;
import java.time.Duration;
import java.time.Instant;
import java.util.UUID;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.transaction.support.TransactionTemplate;

/**
 * Retention deletes only what delivery no longer needs. Each test runs in a transaction that is rolled back, so the
 * relay never sees, and publishes, the rows written here.
 */
class EventRetentionIT extends IntegrationTestSupport {

    @Autowired
    private EventRetention retention;

    @Autowired
    private EventRetentionProperties properties;

    @Autowired
    private TransactionTemplate transactions;

    @Test
    void deletesOutboxRowsPublishedLongerAgoThanTheRetentionAndKeepsUnpublishedOnes() {
        inRolledBackTransaction(() -> {
            Instant now = clock.instant();
            Instant pastRetention = now.minus(properties.publishedOutbox()).minusSeconds(1);
            UUID expired = outboxRow(now.minus(Duration.ofDays(30)), pastRetention);
            UUID recent = outboxRow(now.minus(Duration.ofHours(2)), now.minus(Duration.ofHours(1)));
            UUID unpublished = outboxRow(now.minus(Duration.ofDays(30)), null);

            retention.prune();

            assertThat(outboxRowExists(expired)).isFalse();
            assertThat(outboxRowExists(recent)).isTrue();
            assertThat(outboxRowExists(unpublished)).isTrue();
        });
    }

    @Test
    void deletesDeDuplicationMarkersOlderThanTheRetention() {
        inRolledBackTransaction(() -> {
            Instant now = clock.instant();
            UUID expired = marker(now.minus(properties.processedEvents()).minusSeconds(1));
            UUID recent = marker(now.minus(properties.processedEvents()).plus(Duration.ofHours(1)));

            retention.prune();

            assertThat(markerExists(expired)).isFalse();
            assertThat(markerExists(recent)).isTrue();
        });
    }

    @Test
    void deletesABacklogLargerThanOneBatch() {
        inRolledBackTransaction(() -> {
            Instant publishedAt = clock.instant().minus(properties.publishedOutbox()).minus(Duration.ofDays(1));
            String aggregateId = "retention-" + UUID.randomUUID();
            int rows = properties.batchSize() * 2 + 1;
            jdbc.sql("""
                            INSERT INTO outbox (event_id, topic, message_key, aggregate_id, event_type, payload,
                                                created_at, published_at)
                            SELECT gen_random_uuid(), 'ride-events', :aggregateId, :aggregateId, 'RideRequested',
                                   CAST('{}' AS jsonb), :publishedAt, :publishedAt
                            FROM generate_series(1, :rows)
                            """)
                    .param("aggregateId", aggregateId)
                    .param("publishedAt", Timestamps.toDb(publishedAt))
                    .param("rows", rows)
                    .update();

            retention.prune();

            assertThat(jdbc.sql("SELECT count(*) FROM outbox WHERE aggregate_id = :aggregateId")
                    .param("aggregateId", aggregateId).query(Long.class).single()).isZero();
        });
    }

    private UUID outboxRow(Instant createdAt, Instant publishedAt) {
        UUID eventId = UUID.randomUUID();
        jdbc.sql("""
                        INSERT INTO outbox (event_id, topic, message_key, aggregate_id, event_type, payload,
                                            created_at, published_at)
                        VALUES (:eventId, 'ride-events', :key, :key, 'RideRequested', CAST('{}' AS jsonb),
                                :createdAt, :publishedAt)
                        """)
                .param("eventId", eventId)
                .param("key", eventId.toString())
                .param("createdAt", Timestamps.toDb(createdAt))
                .param("publishedAt", publishedAt == null ? null : Timestamps.toDb(publishedAt))
                .update();
        return eventId;
    }

    private UUID marker(Instant processedAt) {
        UUID eventId = UUID.randomUUID();
        jdbc.sql("""
                        INSERT INTO processed_events (consumer, event_id, processed_at)
                        VALUES ('retention-it', :eventId, :at)
                        """)
                .param("eventId", eventId)
                .param("at", Timestamps.toDb(processedAt))
                .update();
        return eventId;
    }

    private boolean outboxRowExists(UUID eventId) {
        return jdbc.sql("SELECT count(*) FROM outbox WHERE event_id = :eventId")
                .param("eventId", eventId).query(Long.class).single() == 1;
    }

    private boolean markerExists(UUID eventId) {
        return jdbc.sql("SELECT count(*) FROM processed_events WHERE event_id = :eventId")
                .param("eventId", eventId).query(Long.class).single() == 1;
    }

    private void inRolledBackTransaction(Runnable work) {
        transactions.executeWithoutResult(status -> {
            status.setRollbackOnly();
            work.run();
        });
    }
}
