package com.ridehailing.contracts;

import java.util.UUID;

/**
 * The driver rejected; the ride is back to matching and this driver is excluded from it.
 */
public record OfferRejectedV1(UUID rideId, String driverId, UUID offerId) {
}
