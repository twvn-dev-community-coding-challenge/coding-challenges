package com.ridehailing.dispatch.application;

import java.time.Duration;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.boot.context.properties.bind.DefaultValue;

@ConfigurationProperties("dispatch")
public record DispatchProperties(
        @DefaultValue("2.0") double maxRadiusKm,
        @DefaultValue("30s") Duration offerTimeout,
        @DefaultValue("10s") Duration sweeperStaleAfter,
        @DefaultValue("50") int sweeperBatchSize) {
}
