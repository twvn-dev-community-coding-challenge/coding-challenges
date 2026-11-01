package com.ridehailing.ride.application;

import com.ridehailing.common.error.ConflictException;
import com.ridehailing.common.events.OutboxWriter;
import com.ridehailing.common.kernel.TransitionOutcome;
import com.ridehailing.common.logging.LogContext;
import com.ridehailing.contracts.EventType;
import com.ridehailing.contracts.RideCancelledV1;
import com.ridehailing.contracts.RideCompletedV1;
import com.ridehailing.ride.domain.Ride;
import com.ridehailing.ride.domain.RideTransition;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.Map;
import java.util.UUID;

/**
 * Passenger and trip commands the ride owns. They change only the ride and announce it; dispatch releases the driver
 * and cancels any pending offer when it consumes the event. A terminal ride wins every race: dispatch events that
 * arrive after cancellation are ignored by the projection.
 */
@Service
@RequiredArgsConstructor
@Slf4j
public class RideCommandService {

    private final RideLifecycle rides;
    private final OutboxWriter outbox;

    @Transactional
    public Ride cancel(UUID rideId) {
        try (LogContext ignored = LogContext.ride(rideId)) {
            Ride ride = lock(rideId);
            TransitionOutcome outcome = RideTransition.CANCEL.evaluate(ride.status());
            if (outcome == TransitionOutcome.ALREADY_APPLIED) {
                log.info("Duplicate cancellation ignored");
                return ride;
            }
            if (outcome == TransitionOutcome.REJECTED) {
                throw invalidTransition(ride, RideTransition.CANCEL);
            }
            Ride cancelled = rides.apply(rideId, RideTransition.CANCEL, null, "cancelled")
                    .orElseThrow(() -> new IllegalStateException("Locked ride could not be cancelled"));
            outbox.appendForRide(EventType.RIDE_CANCELLED, rideId, ride.assignedDriverId(),
                    new RideCancelledV1(rideId, ride.status().name(), ride.assignedDriverId()));
            return cancelled;
        }
    }

    @Transactional
    public Ride complete(UUID rideId) {
        try (LogContext ignored = LogContext.ride(rideId)) {
            Ride ride = lock(rideId);
            TransitionOutcome outcome = RideTransition.COMPLETE.evaluate(ride.status());
            if (outcome == TransitionOutcome.ALREADY_APPLIED) {
                log.info("Duplicate completion ignored");
                return ride;
            }
            if (outcome == TransitionOutcome.REJECTED) {
                throw invalidTransition(ride, RideTransition.COMPLETE);
            }
            Ride completed = rides.apply(rideId, RideTransition.COMPLETE, ride.assignedDriverId(), "completed")
                    .orElseThrow(() -> new IllegalStateException("Locked ride could not be completed"));
            outbox.appendForRide(EventType.RIDE_COMPLETED, rideId, ride.assignedDriverId(),
                    new RideCompletedV1(rideId, ride.assignedDriverId()));
            return completed;
        }
    }

    private Ride lock(UUID rideId) {
        return rides.lock(rideId).orElseThrow(() -> RideLifecycle.notFound(rideId));
    }

    private static ConflictException invalidTransition(Ride ride, RideTransition transition) {
        return new ConflictException("INVALID_TRANSITION",
                "Cannot %s a ride in status %s".formatted(transition, ride.status()),
                Map.of("currentStatus", ride.status().name()));
    }
}
