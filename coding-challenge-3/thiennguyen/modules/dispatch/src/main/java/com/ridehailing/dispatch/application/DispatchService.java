package com.ridehailing.dispatch.application;

import com.ridehailing.common.kernel.geo.BoundingBox;
import com.ridehailing.common.kernel.geo.Location;
import com.ridehailing.common.logging.LogContext;
import com.ridehailing.contracts.GeoPoint;
import com.ridehailing.contracts.RideRequestedV1;
import com.ridehailing.dispatch.application.port.DispatchRequestRepository;
import com.ridehailing.dispatch.application.port.DriverSupplyRepository;
import com.ridehailing.dispatch.application.port.ExclusionRepository;
import com.ridehailing.dispatch.application.port.OfferRepository;
import com.ridehailing.dispatch.domain.AssignmentTransition;
import com.ridehailing.dispatch.domain.DispatchRequest;
import com.ridehailing.dispatch.domain.DispatchRequestTransition;
import com.ridehailing.dispatch.domain.RideOffer;
import com.ridehailing.dispatch.domain.SupplyChange;
import com.ridehailing.dispatch.domain.SupplyDriver;
import com.ridehailing.dispatch.domain.matching.DriverCandidate;
import com.ridehailing.dispatch.domain.matching.NearestDriverMatcher;
import com.ridehailing.dispatch.domain.matching.RankedCandidate;
import java.time.Clock;
import java.time.Duration;
import java.time.Instant;
import java.util.List;
import java.util.Optional;
import java.util.Set;
import java.util.UUID;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

/**
 * Opens a dispatch request for each requested ride and runs matching attempts. An attempt is one local transaction on
 * dispatch's own tables: lock the request, rank free drivers from dispatch's supply copy, compare-and-set the first one
 * still free, create the pending offer, and publish RideOffered.
 */
@Service
@RequiredArgsConstructor
@Slf4j
public class DispatchService {

    private final DispatchRequestRepository requests;
    private final DriverSupplyRepository supply;
    private final OfferRepository offers;
    private final ExclusionRepository exclusions;
    private final NearestDriverMatcher matcher;
    private final DispatchEvents events;
    private final DispatchProperties properties;
    private final DispatchMetrics metrics;
    private final Clock clock;

    /** RideRequested: start serving the ride and make the first matching attempt in the same transaction. */
    @Transactional
    public DispatchOutcome open(RideRequestedV1 requested) {
        UUID rideId = requested.rideId();
        try (LogContext ignored = LogContext.ride(rideId)) {
            GeoPoint pickup = requested.pickup();
            if (!requests.open(rideId, new Location(pickup.latitude(), pickup.longitude()))) {
                log.info("Dispatch request already open, or the ride already ended; nothing to do");
                return DispatchOutcome.SKIPPED;
            }
            events.matchingStarted(rideId);
            log.info("Dispatch request opened");
            return dispatch(rideId);
        }
    }

    @Transactional
    public DispatchOutcome dispatch(UUID rideId) {
        try (LogContext ignored = LogContext.ride(rideId)) {
            Optional<DispatchRequest> waiting = requests.lockForMatching(rideId);
            if (waiting.isEmpty()) {
                log.info("Dispatch skipped: ride is not waiting for a driver");
                return DispatchOutcome.SKIPPED;
            }
            DispatchRequest request = waiting.get();
            Set<String> excluded = exclusions.forRide(rideId);
            List<DriverCandidate> nearby = supply
                    .findFreeWithin(BoundingBox.around(request.pickup(), matcher.maxRadiusKm())).stream()
                    .map(SupplyDriver::toCandidate)
                    .toList();
            for (RankedCandidate candidate : matcher.rank(request.pickup(), nearby, excluded)) {
                Optional<SupplyChange> claimed = supply.apply(candidate.driverId(), rideId, AssignmentTransition.CLAIM);
                if (claimed.isPresent()) {
                    offer(request, candidate, claimed.get(), !excluded.isEmpty());
                    metrics.attempt(DispatchOutcome.OFFERED);
                    return DispatchOutcome.OFFERED;
                }
                metrics.claimLost();
                try (LogContext driver = LogContext.with(LogContext.DRIVER_ID, candidate.driverId())) {
                    log.info("Candidate was claimed by another dispatch; trying the next one");
                }
            }
            log.info("No eligible driver within {} km ({} excluded); ride stays MATCHING",
                    matcher.maxRadiusKm(), excluded.size());
            metrics.attempt(DispatchOutcome.NO_DRIVER);
            return DispatchOutcome.NO_DRIVER;
        }
    }

    /** {@code rematch}: an earlier offer of this ride was rejected or expired. */
    private void offer(DispatchRequest request, RankedCandidate candidate, SupplyChange claim, boolean rematch) {
        UUID rideId = request.rideId();
        String driverId = candidate.driverId();
        if (!requests.apply(rideId, DispatchRequestTransition.OFFER, driverId)) {
            throw new IllegalStateException("Locked WAITING request could not be offered");
        }
        Instant now = clock.instant();
        RideOffer offer = RideOffer.pending(rideId, driverId, candidate.distanceKm(), now, properties.offerTimeout());
        offers.insert(offer);
        events.rideOffered(offer);
        events.assignmentChanged(claim, rideId);
        metrics.offered(Duration.between(request.createdAt(), now), rematch);
        try (LogContext ignored = LogContext.with(LogContext.DRIVER_ID, driverId)) {
            log.info("Ride offered to driver {} km away, expires at {}",
                    String.format("%.3f", candidate.distanceKm()), offer.expiresAt());
        }
    }
}
