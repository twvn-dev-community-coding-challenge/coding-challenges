package com.ridehailing.integration;

import static org.assertj.core.api.Assertions.assertThat;

import com.ridehailing.common.events.OutboxBacklog;
import com.ridehailing.common.events.OutboxProperties;
import com.ridehailing.common.events.OutboxRelay;
import com.ridehailing.common.jdbc.Timestamps;
import com.ridehailing.support.IntegrationTestSupport;
import io.micrometer.core.instrument.MeterRegistry;
import java.sql.Connection;
import java.sql.PreparedStatement;
import java.time.Duration;
import java.time.Instant;
import java.util.ArrayList;
import java.util.List;
import java.util.UUID;
import java.util.concurrent.TimeUnit;
import java.util.regex.Matcher;
import java.util.regex.Pattern;
import javax.sql.DataSource;
import org.apache.kafka.clients.consumer.Consumer;
import org.apache.kafka.clients.consumer.ConsumerRecord;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.kafka.config.TopicBuilder;
import org.springframework.kafka.core.ConsumerFactory;
import org.springframework.kafka.core.KafkaAdmin;

/**
 * The relay publishes only while it holds the outbox advisory lock, and a backlog of several batches reaches Kafka in
 * the order it was recorded. Rows go to a topic of their own, so no module consumes them.
 */
class OutboxRelayIT extends IntegrationTestSupport {

    private static final String TOPIC = "outbox-relay-it";
    private static final Pattern SEQUENCE = Pattern.compile("\"seq\":\\s*(\\d+)");

    @Autowired
    private DataSource dataSource;

    @Autowired
    private KafkaAdmin kafkaAdmin;

    @Autowired
    private ConsumerFactory<String, String> consumers;

    @Autowired
    private OutboxProperties properties;

    @Autowired
    private OutboxBacklog backlog;

    @Autowired
    private MeterRegistry meters;

    @BeforeEach
    void createTopic() {
        kafkaAdmin.createOrModifyTopics(TopicBuilder.name(TOPIC).partitions(1).replicas(1).build());
    }

    /** Another instance holding the lock stands in for a second relay: this application's relay must stay idle. */
    @Test
    void publishesNothingWhileAnotherRelayHoldsTheLockAndReportsTheBacklog() throws Exception {
        String key = "locked-" + UUID.randomUUID();
        try (Connection otherInstance = dataSource.getConnection()) {
            advisoryLock(otherInstance, "SELECT pg_advisory_lock(?)");
            try {
                insertRows(key, 3, clock.instant().minus(Duration.ofMinutes(1)));

                eventually().during(Duration.ofSeconds(1)).atMost(Duration.ofSeconds(5))
                        .until(() -> unpublished(key) == 3);
                backlog.measure();
                assertThat(meters.get("ridehailing.outbox.pending").gauge().value()).isGreaterThanOrEqualTo(3);
                assertThat(meters.get("ridehailing.outbox.oldest.age").timeGauge().value(TimeUnit.SECONDS))
                        .isGreaterThanOrEqualTo(60);
            } finally {
                advisoryLock(otherInstance, "SELECT pg_advisory_unlock(?)");
            }
        }

        eventually().until(() -> unpublished(key) == 0);
        assertThat(sequencesFor(key, 3)).containsExactly(1, 2, 3);
    }

    @Test
    void publishesABacklogOfSeveralBatchesInRecordedOrderForItsKey() {
        String key = "ordered-" + UUID.randomUUID();
        int rows = properties.batchSize() * 2 + 7;

        insertRows(key, rows, clock.instant());

        eventually().until(() -> unpublished(key) == 0);
        List<Integer> expected = new ArrayList<>();
        for (int sequence = 1; sequence <= rows; sequence++) {
            expected.add(sequence);
        }
        assertThat(sequencesFor(key, rows)).containsExactlyElementsOf(expected);
    }

    private static void advisoryLock(Connection connection, String sql) throws Exception {
        try (PreparedStatement statement = connection.prepareStatement(sql)) {
            statement.setLong(1, OutboxRelay.LOCK_KEY);
            statement.execute();
        }
    }

    /** One committed insert, in sequence order, so the ids follow the sequence. */
    private void insertRows(String key, int rows, Instant createdAt) {
        jdbc.sql("""
                        INSERT INTO outbox (event_id, topic, message_key, aggregate_id, event_type, payload, created_at)
                        SELECT gen_random_uuid(), :topic, :key, :key, 'RelayTest', jsonb_build_object('seq', n),
                               :createdAt
                        FROM generate_series(1, :rows) AS n
                        ORDER BY n
                        """)
                .param("topic", TOPIC)
                .param("key", key)
                .param("createdAt", Timestamps.toDb(createdAt))
                .param("rows", rows)
                .update();
    }

    private long unpublished(String key) {
        return jdbc.sql("SELECT count(*) FROM outbox WHERE message_key = :key AND published_at IS NULL")
                .param("key", key)
                .query(Long.class)
                .single();
    }

    /** The sequence numbers published under {@code key}, in the order Kafka holds them. */
    private List<Integer> sequencesFor(String key, int expected) {
        List<Integer> sequences = new ArrayList<>();
        try (Consumer<String, String> consumer = consumers.createConsumer("probe-" + UUID.randomUUID(), "probe")) {
            consumer.subscribe(List.of(TOPIC));
            long deadline = System.nanoTime() + Duration.ofSeconds(30).toNanos();
            while (sequences.size() < expected && System.nanoTime() < deadline) {
                for (ConsumerRecord<String, String> record : consumer.poll(Duration.ofMillis(500))) {
                    if (key.equals(record.key())) {
                        Matcher sequence = SEQUENCE.matcher(record.value());
                        assertThat(sequence.find()).as(record.value()).isTrue();
                        sequences.add(Integer.parseInt(sequence.group(1)));
                    }
                }
            }
        }
        return sequences;
    }
}
