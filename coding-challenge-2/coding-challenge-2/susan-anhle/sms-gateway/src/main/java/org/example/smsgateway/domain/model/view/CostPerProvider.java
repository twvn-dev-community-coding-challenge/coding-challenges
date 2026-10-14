package org.example.smsgateway.domain.model.view;

import java.math.BigDecimal;

public class CostPerProvider {
    private String providerId;
    private BigDecimal actualCost;
    private BigDecimal estimatedCost;

    public CostPerProvider(String providerId, BigDecimal estimatedCost, BigDecimal actualCost) {
        this.providerId = providerId;
        this.estimatedCost = estimatedCost;
        this.actualCost = actualCost;
    }


    public void addActualCost(BigDecimal cost) {
        this.actualCost = this.actualCost.add(cost);
    }

    public void addEstimatedCost(BigDecimal cost) {
        this.estimatedCost = this.estimatedCost.add(cost);
    }


    public String getProviderId() {
        return providerId;
    }

    public BigDecimal getEstimatedCost() {
        return estimatedCost;
    }

    public BigDecimal getActualCost() {
        return actualCost;
    }
}
