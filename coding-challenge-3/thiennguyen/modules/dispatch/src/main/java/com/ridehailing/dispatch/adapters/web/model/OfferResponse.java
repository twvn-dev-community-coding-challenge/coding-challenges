package com.ridehailing.dispatch.adapters.web.model;

import com.ridehailing.dispatch.domain.OfferStatus;
import com.ridehailing.dispatch.domain.RideOffer;
import java.time.Instant;
import java.util.UUID;

public record OfferResponse(
        UUID id, UUID rideId, String driverId, OfferStatus status, double distanceKm, Instant createdAt,
        Instant expiresAt, Instant respondedAt) {

    public static OfferResponse from(RideOffer offer) {
        return new OfferResponse(offer.id(), offer.rideId(), offer.driverId(), offer.status(), offer.distanceKm(),
                offer.createdAt(), offer.expiresAt(), offer.respondedAt());
    }
}
