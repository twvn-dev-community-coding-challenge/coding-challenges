package com.ridehailing.contracts;

import java.util.UUID;

/**
 * A passenger requested a ride; carries the pickup so dispatch can match without calling ride.
 */
public record RideRequestedV1(UUID rideId, String passengerId, GeoPoint pickup) {
}
