package com.ridehailing.ride.domain;

import com.ridehailing.common.kernel.geo.Location;

import java.time.Instant;
import java.util.UUID;

/**
 * {@code assignedDriverId} holds the offered driver in DRIVER_ASSIGNED and the accepted driver afterwards.
 */
public record Ride(
        UUID id,
        String passengerId,
        Location pickup,
        Location destination,
        RideStatus status,
        String assignedDriverId,
        long version,
        Instant createdAt,
        Instant updatedAt) {

    /** A new ride, waiting for dispatch: no driver yet, first version. */
    public static Ride requested(UUID id, String passengerId, Location pickup, Location destination, Instant now) {
        return new Ride(id, passengerId, pickup, destination, RideStatus.REQUESTED, null, 0L, now, now);
    }
}
