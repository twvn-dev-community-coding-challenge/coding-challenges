package org.example.smsgateway.domain.model.message;

import java.math.BigDecimal;

public record RoutePlan(
        String providerId, BigDecimal estimatedCost
) {
}
