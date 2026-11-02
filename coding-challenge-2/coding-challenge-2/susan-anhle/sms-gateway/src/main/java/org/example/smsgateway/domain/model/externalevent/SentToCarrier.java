package org.example.smsgateway.domain.model.externalevent;

import org.example.smsgateway.domain.model.common.ExternalEvent;

public record SentToCarrier(String messageId) implements ExternalEvent {

}
