package com.ridehailing.contracts;

import java.util.UUID;

/**
 * Dispatch accepted the ride request and is looking for a driver.
 */
public record MatchingStartedV1(UUID rideId) {
}
