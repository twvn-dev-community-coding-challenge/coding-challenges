package org.example.smsgateway.domain.service;

import java.util.Comparator;
import java.util.Map;

public class CarrierIdResolverImpl implements CarrierIdResolver {

    private final Map<String, String> prefixToCarrierId;

    public CarrierIdResolverImpl(Map<String, String> prefixToCarrierId) {
        this.prefixToCarrierId = prefixToCarrierId;
    }

    @Override
    public String determineCarrierId(String phoneNumber) {
        return prefixToCarrierId.entrySet().stream()
                .filter(entry -> phoneNumber.startsWith(entry.getKey()))
                .max(Comparator.comparingInt(entry -> entry.getKey().length()))
                .map(Map.Entry::getValue)
                .orElseThrow(() -> new IllegalStateException("No carrier found for phone number: " + phoneNumber));
    }
}
