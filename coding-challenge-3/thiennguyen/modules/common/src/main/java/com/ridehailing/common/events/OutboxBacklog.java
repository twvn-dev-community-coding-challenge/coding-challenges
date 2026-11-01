package com.ridehailing.common.events;

import com.ridehailing.common.jdbc.Timestamps;
import io.micrometer.core.instrument.Gauge;
import io.micrometer.core.instrument.MeterRegistry;
import io.micrometer.core.instrument.TimeGauge;
import java.time.Clock;
import java.time.Duration;
import java.time.Instant;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicLong;
import lombok.extern.slf4j.Slf4j;
import org.springframework.jdbc.core.simple.JdbcClient;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;

/**
 * How far {@link OutboxRelay} is behind: {@code ridehailing.outbox.pending} counts unpublished rows and
 * {@code ridehailing.outbox.oldest.age} is how long the oldest of them has waited. Measured on its own schedule, so the
 * gauges keep moving while the relay is stuck or Kafka is unreachable.
 */
@Component
@Slf4j
public class OutboxBacklog {

    private final JdbcClient jdbc;
    private final Clock clock;
    private final AtomicLong pending = new AtomicLong();
    private final AtomicLong oldestAgeMillis = new AtomicLong();

    public OutboxBacklog(JdbcClient jdbc, Clock clock, MeterRegistry meters) {
        this.jdbc = jdbc;
        this.clock = clock;
        Gauge.builder("ridehailing.outbox.pending", pending, AtomicLong::get)
                .description("Outbox rows not yet published to Kafka")
                .register(meters);
        TimeGauge.builder("ridehailing.outbox.oldest.age", oldestAgeMillis, TimeUnit.MILLISECONDS, AtomicLong::get)
                .description("How long the oldest unpublished outbox row has waited")
                .register(meters);
    }

    @Scheduled(fixedDelayString = "${outbox.backlog.interval-ms:5000}")
    public void measure() {
        try {
            jdbc.sql("SELECT count(*) AS pending, min(created_at) AS oldest FROM outbox WHERE published_at IS NULL")
                    .query(resultSet -> {
                        Instant oldest = Timestamps.fromDb(resultSet, "oldest");
                        pending.set(resultSet.getLong("pending"));
                        oldestAgeMillis.set(oldest == null
                                ? 0
                                : Math.max(0, Duration.between(oldest, clock.instant()).toMillis()));
                    });
        } catch (RuntimeException exception) {
            log.warn("Could not measure the outbox backlog; the gauges keep their last value: {}",
                    exception.toString());
        }
    }
}
