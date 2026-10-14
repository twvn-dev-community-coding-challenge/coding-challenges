package org.example.smsgateway.domain.model.externalevent;

import org.example.smsgateway.domain.model.common.ExternalEvent;

public record CarrierRejected(String messageId) implements ExternalEvent {
    @Override
    public String messageId() {
        return messageId;
    }
}
