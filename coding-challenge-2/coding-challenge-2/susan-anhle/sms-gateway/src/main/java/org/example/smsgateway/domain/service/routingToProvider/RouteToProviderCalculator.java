package org.example.smsgateway.domain.service.routingToProvider;

import org.example.smsgateway.domain.model.message.RoutePlan;

public interface RouteToProviderCalculator {
    RoutePlan calculateRoute(String country, String carrierId);

}
