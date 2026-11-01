package com.ridehailing.ride.application.port;

import com.ridehailing.ride.domain.Ride;

import java.util.Optional;
import java.util.UUID;

/**
 * Read model for polling clients; a newer version always wins and the cache never decides a change.
 */
public interface RideReadCache {

    Optional<Ride> get(UUID rideId);

    void put(Ride ride);

    /**
     * Stores the ride once the surrounding transaction commits; a rolled-back change never reaches the cache.
     */
    void putAfterCommit(Ride ride);
}
