package org.example.smsgateway.domain.model.view;

import java.math.BigDecimal;
import java.util.Objects;

public class CostPerCountry {
  private String country;
  private BigDecimal estimatedCost;
  private BigDecimal actualCost;

  public CostPerCountry(String country, BigDecimal estimatedCost, BigDecimal actualCost) {
    this.country = country;
    this.estimatedCost = estimatedCost;
    this.actualCost = actualCost;
  }

  public void addEstimatedCost(BigDecimal cost) {
    if (cost == null) cost = BigDecimal.ZERO;
    this.estimatedCost = this.estimatedCost.add(cost);
  }

  public void addActualCost(BigDecimal cost) {
    if (cost == null) cost = BigDecimal.ZERO;
    this.actualCost = this.actualCost.add(cost);
  }

  public String getCountry() {
    return country;
  }

  public BigDecimal getEstimatedCost() {
    return estimatedCost;
  }

  public BigDecimal getActualCost() {
    return actualCost;
  }

  @Override
  public boolean equals(Object o) {
    if (this == o) return true;
    if (o == null || getClass() != o.getClass()) return false;
    CostPerCountry that = (CostPerCountry) o;
    return Objects.equals(country, that.country)
        && Objects.equals(actualCost, that.actualCost)
        && Objects.equals(estimatedCost, that.estimatedCost);
  }

  @Override
  public int hashCode() {
    return Objects.hash(country, actualCost, estimatedCost);
  }
}
