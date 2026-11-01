package com.ridehailing.dispatch.application;

import com.ridehailing.common.events.OutboxWriter;
import com.ridehailing.contracts.AssignmentState;
import com.ridehailing.contracts.DriverAssignmentChangedV1;
import com.ridehailing.contracts.EventType;
import com.ridehailing.contracts.MatchingStartedV1;
import com.ridehailing.contracts.OfferExpiredV1;
import com.ridehailing.contracts.OfferRejectedV1;
import com.ridehailing.contracts.RideAcceptedV1;
import com.ridehailing.contracts.RideOfferedV1;
import com.ridehailing.dispatch.domain.RideOffer;
import com.ridehailing.dispatch.domain.SupplyChange;
import java.util.UUID;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Component;

/** Everything dispatch publishes on dispatch-events, written to the outbox in the caller's transaction. */
@Component
@RequiredArgsConstructor
public class DispatchEvents {

    private final OutboxWriter outbox;

    public void matchingStarted(UUID rideId) {
        outbox.appendForRide(EventType.MATCHING_STARTED, rideId, null, new MatchingStartedV1(rideId));
    }

    public void rideOffered(RideOffer offer) {
        outbox.appendForRide(EventType.RIDE_OFFERED, offer.rideId(), offer.driverId(), new RideOfferedV1(
                offer.rideId(), offer.driverId(), offer.id(), offer.distanceKm(), offer.expiresAt()));
    }

    public void offerRejected(RideOffer offer) {
        outbox.appendForRide(EventType.OFFER_REJECTED, offer.rideId(), offer.driverId(),
                new OfferRejectedV1(offer.rideId(), offer.driverId(), offer.id()));
    }

    public void offerExpired(RideOffer offer, String reason) {
        outbox.appendForRide(EventType.OFFER_EXPIRED, offer.rideId(), offer.driverId(),
                new OfferExpiredV1(offer.rideId(), offer.driverId(), offer.id(), reason));
    }

    public void rideAccepted(RideOffer offer) {
        outbox.appendForRide(EventType.RIDE_ACCEPTED, offer.rideId(), offer.driverId(),
                new RideAcceptedV1(offer.rideId(), offer.driverId(), offer.id()));
    }

    public void assignmentChanged(SupplyChange change, UUID rideId) {
        outbox.appendForDriver(EventType.DRIVER_ASSIGNMENT_CHANGED, change.driverId(), rideId,
                new DriverAssignmentChangedV1(change.driverId(), change.rideId(),
                        AssignmentState.valueOf(change.assignment().name()), change.version()));
    }
}
