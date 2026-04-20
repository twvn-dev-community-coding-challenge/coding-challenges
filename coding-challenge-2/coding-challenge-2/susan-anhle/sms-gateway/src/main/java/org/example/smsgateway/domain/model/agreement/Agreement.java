package org.example.smsgateway.domain.model.agreement;

import java.math.BigDecimal;

public class Agreement {
    private Carrier carrier;
    private Provider provider;
    private String country;
    private BigDecimal price;

    public Agreement(Carrier carrier, Provider provider, String country, BigDecimal price) {
        this.carrier = carrier;
        this.provider = provider;
        this.country = country;
        this.price = price;
    }

    public String getCarrierId() {
        return carrier.getCarrierId();
    }

    public String getCountry() {
        return country;
    }

    public String getProviderId() {
        return provider.getId();
    }

    public BigDecimal getPrice() {
        return price;
    }
}