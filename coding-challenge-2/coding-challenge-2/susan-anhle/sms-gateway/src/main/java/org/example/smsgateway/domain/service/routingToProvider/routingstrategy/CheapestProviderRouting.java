package org.example.smsgateway.domain.service.routingToProvider.routingstrategy;

import org.example.smsgateway.domain.model.agreement.Agreement;
import org.example.smsgateway.domain.model.agreement.AgreementRepository;
import org.example.smsgateway.domain.model.message.RoutePlan;
import org.example.smsgateway.domain.service.routingToProvider.RouteToProviderCalculator;

public class CheapestProviderRouting implements RouteToProviderCalculator {
    private final AgreementRepository agreementRepository;

    public CheapestProviderRouting(AgreementRepository agreementRepository) {
        this.agreementRepository = agreementRepository;
    }

    @Override
    public RoutePlan calculateRoute(String country, String carrierId) {
        Agreement agreement = agreementRepository.getCheapestAgreement(carrierId, country);
        return new RoutePlan(agreement.getProviderId(), agreement.getPrice());
    }
}
