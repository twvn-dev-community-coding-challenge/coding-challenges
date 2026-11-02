package org.example.smsgateway.domain.model.message.event.smsStatus;

import org.example.smsgateway.domain.model.message.event.MessageEvent;

public record Queued(String messageId) implements MessageEvent {
}
