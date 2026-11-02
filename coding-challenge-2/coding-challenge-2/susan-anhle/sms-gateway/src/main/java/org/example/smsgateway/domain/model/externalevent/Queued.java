package org.example.smsgateway.domain.model.externalevent;

import org.example.smsgateway.domain.model.common.ExternalEvent;

import java.math.BigDecimal;

public record Queued(String messageId) implements ExternalEvent {

}
