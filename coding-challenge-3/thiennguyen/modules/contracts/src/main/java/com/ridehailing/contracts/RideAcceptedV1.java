package com.ridehailing.contracts;

import java.util.UUID;

public record RideAcceptedV1(UUID rideId, String driverId, UUID offerId) {
}
