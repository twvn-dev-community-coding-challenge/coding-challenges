package com.ridehailing.ride.application;

import com.ridehailing.common.error.NotFoundException;
import com.ridehailing.common.logging.LogContext;
import com.ridehailing.ride.application.port.RideReadCache;
import com.ridehailing.ride.application.port.RideRepository;
import com.ridehailing.ride.domain.Ride;
import com.ridehailing.ride.domain.RideHistoryEntry;
import com.ridehailing.ride.domain.RideTransition;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Propagation;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

/**
 * The ride module's state: reads for the API, and compare-and-set transitions that write history and refresh the read
 * cache after commit. Only the ride module changes ride state — its own commands (create, cancel, complete) and its
 * projection of dispatch events. Transitions must run inside the caller's transaction.
 */
@Service
@RequiredArgsConstructor
@Slf4j
public class RideLifecycle {

    private final RideRepository rides;
    private final RideReadCache cache;

    /**
     * Cache-first read for polling clients; never use the result to decide a state change.
     */
    public Ride getForRead(UUID rideId) {
        return cache.get(rideId).orElseGet(() -> {
            Ride ride = rides.find(rideId).orElseThrow(() -> notFound(rideId));
            cache.put(ride);
            return ride;
        });
    }

    public Optional<Ride> find(UUID rideId) {
        return rides.find(rideId);
    }

    @Transactional(propagation = Propagation.MANDATORY)
    public Optional<Ride> lock(UUID rideId) {
        return rides.findForUpdate(rideId);
    }

    @Transactional(propagation = Propagation.MANDATORY)
    public Optional<Ride> apply(UUID rideId, RideTransition transition, String driverId, String reason) {
        Optional<RideRepository.TransitionResult> result = rides.transition(rideId, transition, driverId);
        result.ifPresent(change -> {
            Ride ride = change.ride();
            String historyDriver = driverId != null ? driverId : ride.assignedDriverId();
            rides.recordHistory(rideId, change.previousStatus(), ride.status(), historyDriver, reason);
            cache.putAfterCommit(ride);
            try (LogContext ignored = LogContext.rideAndDriver(rideId, historyDriver)) {
                log.info("Ride {} -> {} ({})", change.previousStatus(), ride.status(), reason);
            }
        });
        return result.map(RideRepository.TransitionResult::ride);
    }

    public List<RideHistoryEntry> history(UUID rideId) {
        if (rides.find(rideId).isEmpty()) {
            throw notFound(rideId);
        }
        return rides.history(rideId);
    }

    public static NotFoundException notFound(UUID rideId) {
        return new NotFoundException("RIDE_NOT_FOUND", "Ride %s does not exist".formatted(rideId));
    }
}
