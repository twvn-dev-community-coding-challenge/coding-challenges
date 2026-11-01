package com.ridehailing.contracts;

import java.util.UUID;

/**
 * The accepted trip finished; terminal.
 */
public record RideCompletedV1(UUID rideId, String driverId) {
}
