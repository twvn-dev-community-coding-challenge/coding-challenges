package com.ridehailing.common.cache;

import java.time.Duration;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.boot.context.properties.bind.DefaultValue;

@ConfigurationProperties("cache")
public record CacheProperties(
        @DefaultValue("30s") Duration rideTtl,
        @DefaultValue("30s") Duration driverTtl,
        @DefaultValue("24h") Duration idempotencyTtl) {
}
