package com.ridehailing.dispatch.application;

import com.ridehailing.common.error.ConflictException;
import com.ridehailing.common.error.NotFoundException;
import com.ridehailing.common.kernel.TransitionOutcome;
import com.ridehailing.common.logging.LogContext;
import com.ridehailing.contracts.OfferExpiredV1;
import com.ridehailing.dispatch.application.DispatchMetrics.OfferEnd;
import com.ridehailing.dispatch.application.port.DispatchRequestRepository;
import com.ridehailing.dispatch.application.port.DriverSupplyRepository;
import com.ridehailing.dispatch.application.port.ExclusionRepository;
import com.ridehailing.dispatch.application.port.OfferRepository;
import com.ridehailing.dispatch.domain.AssignmentTransition;
import com.ridehailing.dispatch.domain.DispatchRequest;
import com.ridehailing.dispatch.domain.DispatchRequestTransition;
import com.ridehailing.dispatch.domain.OfferStatus;
import com.ridehailing.dispatch.domain.OfferTransition;
import com.ridehailing.dispatch.domain.RideOffer;
import com.ridehailing.dispatch.domain.SupplyChange;
import java.time.Clock;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

/**
 * Driver responses to offers: accept, reject, and expiry by timeout or by the driver going offline. Each is one
 * transaction on dispatch's tables that locks the request row first and the supply row second; the request decides
 * whether the response is allowed ({@link DispatchRequest#evaluate}). Repeating a response that already took effect
 * returns the offer without side effects. The ride learns of the outcome from the published event.
 */
@Service
@RequiredArgsConstructor
@Slf4j
public class OfferService {

    private final DispatchRequestRepository requests;
    private final DriverSupplyRepository supply;
    private final OfferRepository offers;
    private final ExclusionRepository exclusions;
    private final DispatchEvents events;
    private final DispatchMetrics metrics;
    private final Clock clock;

    @Transactional
    public RideOffer accept(UUID rideId, String driverId) {
        try (LogContext ignored = LogContext.rideAndDriver(rideId, driverId)) {
            DispatchRequest request = lock(rideId);
            TransitionOutcome outcome = request.evaluate(DispatchRequestTransition.ACCEPT, driverId);
            if (outcome == TransitionOutcome.ALREADY_APPLIED) {
                log.info("Duplicate acceptance ignored; ride already accepted by this driver");
                return latestOffer(rideId, driverId);
            }
            if (outcome == TransitionOutcome.REJECTED) {
                throw offerNotHeld(request, driverId);
            }
            if (!offers.accept(rideId, driverId, clock.instant())) {
                throw new ConflictException("OFFER_EXPIRED", "The offer for ride %s has expired".formatted(rideId),
                        Map.of("currentStatus", request.status().name()));
            }
            if (!requests.apply(rideId, DispatchRequestTransition.ACCEPT, driverId)) {
                throw new IllegalStateException("Locked OFFERED request could not be accepted");
            }
            SupplyChange busy = supply.apply(driverId, rideId, AssignmentTransition.START_TRIP)
                    .orElseThrow(() -> new IllegalStateException("Offered driver could not start the trip"));
            RideOffer accepted = latestOffer(rideId, driverId);
            events.rideAccepted(accepted);
            events.assignmentChanged(busy, rideId);
            metrics.offerEnded(OfferEnd.ACCEPTED);
            log.info("Offer accepted");
            return accepted;
        }
    }

    @Transactional
    public RideOffer reject(UUID rideId, String driverId) {
        try (LogContext ignored = LogContext.rideAndDriver(rideId, driverId)) {
            DispatchRequest request = lock(rideId);
            if (request.evaluate(DispatchRequestTransition.RETURN_TO_WAITING, driverId) == TransitionOutcome.APPLY) {
                if (!offers.resolvePending(rideId, driverId, OfferTransition.REJECT, clock.instant())) {
                    throw new IllegalStateException("Offered request has no pending offer for its driver");
                }
                RideOffer rejected = latestOffer(rideId, driverId);
                returnToWaiting(rejected, OfferStatus.REJECTED);
                events.offerRejected(rejected);
                metrics.offerEnded(OfferEnd.REJECTED);
                log.info("Offer rejected; driver excluded from this ride");
                return rejected;
            }
            if (exclusions.contains(rideId, driverId)) {
                log.info("Duplicate rejection ignored; driver is already excluded from this ride");
                return latestOffer(rideId, driverId);
            }
            throw offerNotHeld(request, driverId);
        }
    }

    /**
     * An unanswered offer, or one whose driver went offline, is treated as a rejection. Safe to run concurrently with
     * accept, reject and ride cancellation: whichever locks the request first wins.
     */
    @Transactional
    public void expire(RideOffer offer, String reason) {
        try (LogContext ignored = LogContext.rideAndDriver(offer.rideId(), offer.driverId())) {
            DispatchRequest request = lock(offer.rideId());
            if (request.evaluate(DispatchRequestTransition.RETURN_TO_WAITING, offer.driverId())
                    != TransitionOutcome.APPLY) {
                log.info("Expiry skipped: the offer was already answered or the ride moved on");
                return;
            }
            boolean expired = OfferExpiredV1.DRIVER_OFFLINE.equals(reason)
                    ? offers.resolvePending(offer.rideId(), offer.driverId(), OfferTransition.EXPIRE, clock.instant())
                    : offers.expireIfDue(offer.id(), clock.instant());
            if (!expired) {
                log.info("Expiry skipped: the offer is no longer pending or not yet due");
                return;
            }
            RideOffer lapsed = latestOffer(offer.rideId(), offer.driverId());
            returnToWaiting(lapsed, OfferStatus.EXPIRED);
            events.offerExpired(lapsed, reason);
            metrics.offerEnded(
                    OfferExpiredV1.DRIVER_OFFLINE.equals(reason) ? OfferEnd.DRIVER_OFFLINE : OfferEnd.TIMED_OUT);
            log.info("Offer expired ({}); driver excluded from this ride", reason);
        }
    }

    public List<RideOffer> offersFor(UUID rideId) {
        if (requests.find(rideId).isEmpty()) {
            throw notFound(rideId);
        }
        return offers.findByRide(rideId);
    }

    private void returnToWaiting(RideOffer offer, OfferStatus reason) {
        exclusions.add(offer.rideId(), offer.driverId(), reason, clock.instant());
        if (!requests.apply(offer.rideId(), DispatchRequestTransition.RETURN_TO_WAITING, offer.driverId())) {
            throw new IllegalStateException("Locked OFFERED request could not return to waiting");
        }
        SupplyChange released = supply.apply(offer.driverId(), offer.rideId(), AssignmentTransition.RELEASE)
                .orElseThrow(() -> new IllegalStateException("Offered driver could not be released"));
        events.assignmentChanged(released, offer.rideId());
    }

    private DispatchRequest lock(UUID rideId) {
        return requests.lock(rideId).orElseThrow(() -> notFound(rideId));
    }

    private RideOffer latestOffer(UUID rideId, String driverId) {
        return offers.findLatest(rideId, driverId)
                .orElseThrow(() -> new IllegalStateException("No offer of ride %s to %s".formatted(rideId, driverId)));
    }

    static NotFoundException notFound(UUID rideId) {
        return new NotFoundException("RIDE_NOT_FOUND", "Dispatch has no ride %s".formatted(rideId));
    }

    private static ConflictException offerNotHeld(DispatchRequest request, String driverId) {
        return new ConflictException("OFFER_NOT_HELD",
                "Driver %s does not hold the current offer for ride %s".formatted(driverId, request.rideId()),
                Map.of("currentStatus", request.status().name()));
    }
}
