package org.example.smsgateway.domain.model.view;

import java.math.BigDecimal;

class MessageCostView {
    String messageId;
    BigDecimal estimatedCost;
    BigDecimal actualCost;
    String providerId;
    String country;
}
