package com.ridehailing.contracts;

import java.util.UUID;

/**
 * The offer lapsed without an answer; treated exactly like a rejection. {@code reason} is {@code TIMEOUT} or
 * {@code DRIVER_OFFLINE} (the driver went offline while holding the offer).
 */
public record OfferExpiredV1(UUID rideId, String driverId, UUID offerId, String reason) {

    public static final String TIMEOUT = "TIMEOUT";
    public static final String DRIVER_OFFLINE = "DRIVER_OFFLINE";
}
