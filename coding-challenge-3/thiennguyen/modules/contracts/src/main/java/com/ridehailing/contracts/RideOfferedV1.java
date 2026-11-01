package com.ridehailing.contracts;

import java.time.Instant;
import java.util.UUID;

public record RideOfferedV1(UUID rideId, String driverId, UUID offerId, double distanceKm, Instant expiresAt) {
}
