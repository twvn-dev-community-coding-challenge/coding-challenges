package com.ridehailing.ride.application.port;

import com.ridehailing.ride.domain.Ride;
import com.ridehailing.ride.domain.RideHistoryEntry;
import com.ridehailing.ride.domain.RideStatus;
import com.ridehailing.ride.domain.RideTransition;

import java.time.Instant;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

/**
 * The ride aggregate's store. Every status change is a compare-and-set guarded by the {@link RideTransition}.
 */
public interface RideRepository {

    void insert(Ride ride);

    Optional<Ride> find(UUID rideId);

    /**
     * Locks the ride for the rest of the caller's transaction.
     */
    Optional<Ride> findForUpdate(UUID rideId);

    /**
     * Rides still REQUESTED whose RideRequested was last published before {@code requestedBefore}: dispatch never
     * picked them up. Marks them as requested again now and returns them, skipping rides another transaction holds
     * (a cancellation, or the sweeper on another instance). Not a state change: version and history are untouched.
     */
    List<Ride> claimUndispatched(Instant requestedBefore, int limit);

    /**
     * Applies {@code transition} if the ride is in one of its {@link RideTransition#from() from} statuses (and, when
     * {@link RideTransition#requiresAssignedDriver()}, assigned to {@code driverId}). Empty when it was not.
     */
    Optional<TransitionResult> transition(UUID rideId, RideTransition transition, String driverId);

    void recordHistory(UUID rideId, RideStatus from, RideStatus to, String driverId, String reason);

    List<RideHistoryEntry> history(UUID rideId);

    record TransitionResult(Ride ride, RideStatus previousStatus) {
    }
}
