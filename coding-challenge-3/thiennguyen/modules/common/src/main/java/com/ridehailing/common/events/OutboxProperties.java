package com.ridehailing.common.events;

import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.boot.context.properties.bind.DefaultValue;

@ConfigurationProperties("outbox.relay")
public record OutboxProperties(@DefaultValue("100") int batchSize) {
}
