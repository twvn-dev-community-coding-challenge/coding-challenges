package org.example.smsgateway.domain.model.agreement;

public interface AgreementRepository {
    void save(Agreement agreement);

    Agreement getCheapestAgreement(String carrierId, String country);
}
