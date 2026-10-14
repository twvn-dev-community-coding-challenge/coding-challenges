package org.example.smsgateway.domain.service;

public interface CarrierIdResolver {

    String determineCarrierId(String phoneNumber);
}
