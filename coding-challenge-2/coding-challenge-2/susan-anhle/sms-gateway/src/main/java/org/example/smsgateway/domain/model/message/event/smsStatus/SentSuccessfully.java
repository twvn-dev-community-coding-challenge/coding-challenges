package org.example.smsgateway.domain.model.message.event.smsStatus;

import java.math.BigDecimal;
import org.example.smsgateway.domain.model.message.event.MessageEvent;

public record SentSuccessfully(String messageId, BigDecimal actualCost) implements MessageEvent {
}
