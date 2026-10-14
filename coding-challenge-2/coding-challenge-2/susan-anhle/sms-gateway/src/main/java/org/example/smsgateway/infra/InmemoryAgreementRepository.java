package org.example.smsgateway.infra;

import tools.jackson.databind.JsonNode;
import tools.jackson.databind.ObjectMapper;
import org.example.smsgateway.domain.model.agreement.Agreement;
import org.example.smsgateway.domain.model.agreement.AgreementRepository;
import org.example.smsgateway.domain.model.agreement.Carrier;
import org.example.smsgateway.domain.model.agreement.Provider;

import java.io.InputStream;
import java.math.BigDecimal;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.List;

public class InmemoryAgreementRepository implements AgreementRepository {

    private final List<Agreement> agreements = new ArrayList<>();

    public InmemoryAgreementRepository() {
        try (InputStream is = getClass().getResourceAsStream("/agreements.json")) {
            JsonNode nodes = new ObjectMapper().readTree(is);
            for (JsonNode node : nodes) {
                agreements.add(new Agreement(
                        new Carrier(node.get("carrierId").asText(), node.get("carrierName").asText()),
                        new Provider(node.get("providerId").asText()),
                        node.get("country").asText(),
                        new BigDecimal(node.get("price").asText())
                ));
            }
        } catch (Exception e) {
            throw new IllegalStateException("Failed to load agreements.json", e);
        }
    }

    @Override
    public void save(Agreement agreement) {
        agreements.add(agreement);
    }

    @Override
    public Agreement getCheapestAgreement(String carrierId, String country) {
        return agreements.stream()
                .filter(a -> a.getCarrierId().equals(carrierId) && a.getCountry().equals(country))
                .min(Comparator.comparing(Agreement::getPrice))
                .orElseThrow(() -> new IllegalStateException(
                        "No agreement found for carrier=" + carrierId + ", country=" + country));
    }
}
