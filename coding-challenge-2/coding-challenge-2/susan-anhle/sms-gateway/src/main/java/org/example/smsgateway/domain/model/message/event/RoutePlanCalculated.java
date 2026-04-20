package org.example.smsgateway.domain.model.message.event;

import java.math.BigDecimal;

public record RoutePlanCalculated(String messageId, String providerId,
                                  BigDecimal estimatedCost) implements MessageEvent {
}
