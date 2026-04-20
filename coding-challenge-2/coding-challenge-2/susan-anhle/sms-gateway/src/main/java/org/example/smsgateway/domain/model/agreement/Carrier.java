package org.example.smsgateway.domain.model.agreement;

public class Carrier{
    private String carrierId;

    public Carrier(String carrierId, String carrierName) {
        this.carrierId = carrierId;
        this.carrierName = carrierName;
    }

    private String carrierName;

    public String getCarrierId() {
        return carrierId;
    }

    public String getCarrierName() {
        return carrierName;
    }
}
