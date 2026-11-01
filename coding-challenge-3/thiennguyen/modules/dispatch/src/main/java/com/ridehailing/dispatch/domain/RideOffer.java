package com.ridehailing.dispatch.domain;

import java.time.Duration;
import java.time.Instant;
import java.util.UUID;

public record RideOffer(
        UUID id,
        UUID rideId,
        String driverId,
        OfferStatus status,
        double distanceKm,
        Instant createdAt,
        Instant expiresAt,
        Instant respondedAt) {

    /** A new offer awaiting the driver's answer, which expires {@code timeout} after {@code now}. */
    public static RideOffer pending(UUID rideId, String driverId, double distanceKm, Instant now, Duration timeout) {
        return new RideOffer(UUID.randomUUID(), rideId, driverId, OfferStatus.PENDING, distanceKm, now,
                now.plus(timeout), null);
    }
}
