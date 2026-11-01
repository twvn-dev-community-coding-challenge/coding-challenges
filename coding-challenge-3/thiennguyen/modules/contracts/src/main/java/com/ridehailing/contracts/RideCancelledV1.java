package com.ridehailing.contracts;

import java.util.UUID;

/**
 * The ride was cancelled; terminal. {@code driverId} is the driver it held, if any.
 */
public record RideCancelledV1(UUID rideId, String previousStatus, String driverId) {
}
