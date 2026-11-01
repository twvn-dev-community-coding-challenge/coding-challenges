package com.ridehailing.ride.application;

import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.boot.context.properties.bind.DefaultValue;

import java.time.Duration;

/**
 * {@code undispatchedAfter}: how long a ride may stay REQUESTED after its RideRequested was published before the
 * event is published again (dispatch normally answers with MatchingStarted within milliseconds).
 */
@ConfigurationProperties("ride")
public record RideProperties(
        @DefaultValue("30s") Duration undispatchedAfter,
        @DefaultValue("50") int undispatchedBatchSize) {
}
