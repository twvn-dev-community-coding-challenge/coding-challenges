package com.ridehailing.dispatch.application;

import com.ridehailing.common.kernel.TransitionOutcome;
import com.ridehailing.common.logging.LogContext;
import com.ridehailing.dispatch.application.DispatchMetrics.OfferEnd;
import com.ridehailing.dispatch.application.port.DispatchRequestRepository;
import com.ridehailing.dispatch.application.port.DriverSupplyRepository;
import com.ridehailing.dispatch.application.port.OfferRepository;
import com.ridehailing.dispatch.domain.AssignmentTransition;
import com.ridehailing.dispatch.domain.DispatchRequest;
import com.ridehailing.dispatch.domain.DispatchRequestStatus;
import com.ridehailing.dispatch.domain.DispatchRequestTransition;
import com.ridehailing.dispatch.domain.OfferTransition;
import java.time.Clock;
import java.util.Optional;
import java.util.UUID;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Propagation;
import org.springframework.transaction.annotation.Transactional;

/**
 * Reacts to the ride ending (RideCancelled, RideCompleted): closes the dispatch request, cancels a pending offer and
 * frees the driver. The ride's terminal state wins every race — an acceptance that slipped in first is undone here.
 * A ride that ends before dispatch opened its request is recorded as ended, so a late RideRequested is ignored.
 */
@Component
@RequiredArgsConstructor
@Slf4j
public class RideOutcomes {

    private final DispatchRequestRepository requests;
    private final DriverSupplyRepository supply;
    private final OfferRepository offers;
    private final DispatchEvents events;
    private final DispatchMetrics metrics;
    private final Clock clock;

    @Transactional(propagation = Propagation.MANDATORY)
    public void rideEnded(UUID rideId, DispatchRequestTransition ending) {
        try (LogContext ignored = LogContext.ride(rideId)) {
            Optional<DispatchRequest> locked = requests.lock(rideId);
            if (locked.isEmpty()) {
                if (requests.recordEnded(rideId, ending.to())) {
                    log.info("Ride ended before dispatch served it; recorded {} so it is never offered", ending.to());
                    return;
                }
                locked = requests.lock(rideId);
            }
            DispatchRequest request = locked.orElseThrow(
                    () -> new IllegalStateException("Dispatch request vanished while recording the ride's end"));
            if (request.evaluate(ending, null) != TransitionOutcome.APPLY) {
                log.info("Ride end already handled");
                return;
            }
            String driverId = request.driverId();
            if (request.status() == DispatchRequestStatus.OFFERED
                    && offers.resolvePending(rideId, driverId, OfferTransition.CANCEL, clock.instant())) {
                metrics.offerEnded(OfferEnd.CANCELLED);
            }
            if (driverId != null) {
                supply.apply(driverId, rideId, AssignmentTransition.RELEASE).ifPresentOrElse(
                        released -> events.assignmentChanged(released, rideId),
                        () -> log.warn("Driver of an active request was not held for this ride"));
            }
            requests.apply(rideId, ending, driverId);
            try (LogContext driver = LogContext.with(LogContext.DRIVER_ID, driverId)) {
                log.info("Dispatch request {} after the ride ended; driver released", ending.to());
            }
        }
    }
}
