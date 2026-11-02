package org.example.smsgateway.domain.model.view;

import java.math.BigDecimal;
import java.math.RoundingMode;

public class DeliveryRatePerProvider {
    private final String providerId;
    private long sent;
    private long succeeded;
    private long failed;

    public DeliveryRatePerProvider(String providerId) {
        this.providerId = providerId;
    }

    public void incrementSent() {
        this.sent++;
    }

    public void incrementSucceeded() {
        this.succeeded++;
    }

    public void incrementFailed() {
        this.failed++;
    }

    public String getProviderId() {
        return providerId;
    }

    public long getSent() {
        return sent;
    }

    public long getSucceeded() {
        return succeeded;
    }

    public long getFailed() {
        return failed;
    }

    public BigDecimal getSuccessRate() {
        if (sent == 0) return BigDecimal.ZERO;
        return BigDecimal.valueOf(succeeded).divide(BigDecimal.valueOf(sent), 4, RoundingMode.HALF_UP);
    }

    public BigDecimal getFailureRate() {
        if (sent == 0) return BigDecimal.ZERO;
        return BigDecimal.valueOf(failed).divide(BigDecimal.valueOf(sent), 4, RoundingMode.HALF_UP);
    }
}
