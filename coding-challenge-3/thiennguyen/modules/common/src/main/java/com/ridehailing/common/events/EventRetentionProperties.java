package com.ridehailing.common.events;

import com.ridehailing.contracts.Topics;

import java.time.Duration;

import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.boot.context.properties.bind.DefaultValue;

/**
 * How long {@link EventRetention} keeps published outbox rows and de-duplication markers. Markers must outlive every
 * possible redelivery, so {@code processedEvents} must be longer than {@link Topics#RETENTION_DAYS}, which every topic and
 * dead-letter topic uses; startup fails otherwise.
 */
@ConfigurationProperties("events.retention")
public record EventRetentionProperties(
        @DefaultValue("1d") Duration publishedOutbox,
        @DefaultValue("14d") Duration processedEvents,
        @DefaultValue("1000") int batchSize) {

    public EventRetentionProperties {
        if (publishedOutbox.isNegative() || publishedOutbox.isZero()
                || processedEvents.isNegative() || processedEvents.isZero()) {
            throw new IllegalArgumentException("events.retention durations must be positive");
        }
        if (processedEvents.compareTo(Topics.RETENTION_DAYS) <= 0) {
            throw new IllegalArgumentException(
                    "events.retention.processed-events (%s) must be longer than the topic retention (%s)"
                            .formatted(processedEvents, Topics.RETENTION_DAYS));
        }
        if (batchSize < 1) {
            throw new IllegalArgumentException("events.retention.batch-size must be at least 1");
        }
    }
}
