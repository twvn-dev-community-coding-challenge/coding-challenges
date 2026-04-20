package org.example.smsgateway.domain.model.message.event;

public record NewMessageRequestReceived(String messageId, String message, String country,
                                        String phoneNumber) implements MessageEvent {
}
