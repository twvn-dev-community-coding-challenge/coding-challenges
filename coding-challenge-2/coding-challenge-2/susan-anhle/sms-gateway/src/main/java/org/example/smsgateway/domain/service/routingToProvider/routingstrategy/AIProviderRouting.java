package org.example.smsgateway.domain.service.routingToProvider.routingstrategy;

import org.example.smsgateway.domain.model.message.RoutePlan;
import org.example.smsgateway.domain.service.routingToProvider.RouteToProviderCalculator;

public class AIProviderRouting implements RouteToProviderCalculator {
    private final AIRoutingModelGateway aiRoutingModelGateway;

    public AIProviderRouting(AIRoutingModelGateway aiRoutingModelGateway) {
        this.aiRoutingModelGateway = aiRoutingModelGateway;
    }

    @Override
    public RoutePlan calculateRoute(String country, String carrierId) {
        // calling some AIModel to decide
        return new RoutePlan(null, null);
    }

    interface AIRoutingModelGateway {
    }
}
