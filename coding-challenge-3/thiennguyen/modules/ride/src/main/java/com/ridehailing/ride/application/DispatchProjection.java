package com.ridehailing.ride.application;

import com.ridehailing.common.events.EventCodec;
import com.ridehailing.common.events.ProcessedEvents;
import com.ridehailing.contracts.*;
import com.ridehailing.ride.domain.RideTransition;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;

import java.util.EnumSet;
import java.util.Set;
import java.util.UUID;

/**
 * Keeps the ride's dispatch-driven statuses (MATCHING, DRIVER_ASSIGNED, DRIVER_ACCEPTED) in step with dispatch, which
 * owns offers. Each event is applied at most once (processed_events) and only as a compare-and-set, so a cancelled or
 * completed ride ignores late dispatch events: the terminal ride state always wins.
 */
@Component
@RequiredArgsConstructor
@Slf4j
public class DispatchProjection {

    public static final String CONSUMER = "ride";
    public static final Set<EventType> HANDLED_EVENTS = EnumSet.of(EventType.MATCHING_STARTED, EventType.RIDE_OFFERED,
            EventType.OFFER_REJECTED, EventType.OFFER_EXPIRED, EventType.RIDE_ACCEPTED);

    private final RideLifecycle rides;
    private final ProcessedEvents processedEvents;
    private final EventCodec codec;

    @Transactional
    public void apply(EventType type, EventEnvelope envelope) {
        if (!processedEvents.markProcessed(CONSUMER, envelope.eventId())) {
            log.info("Duplicate delivery ignored; event was already applied to the ride");
            return;
        }
        switch (type) {
            case MATCHING_STARTED -> {
                MatchingStartedV1 event = codec.payload(envelope, MatchingStartedV1.class);
                applyOrIgnore(event.rideId(), RideTransition.START_MATCHING, null, "dispatch started");
            }
            case RIDE_OFFERED -> {
                RideOfferedV1 event = codec.payload(envelope, RideOfferedV1.class);
                applyOrIgnore(event.rideId(), RideTransition.ASSIGN_DRIVER, event.driverId(), "offer created");
            }
            case OFFER_REJECTED -> {
                OfferRejectedV1 event = codec.payload(envelope, OfferRejectedV1.class);
                applyOrIgnore(event.rideId(), RideTransition.RETURN_TO_MATCHING, event.driverId(), "driver rejected");
            }
            case OFFER_EXPIRED -> {
                OfferExpiredV1 event = codec.payload(envelope, OfferExpiredV1.class);
                String reason = OfferExpiredV1.DRIVER_OFFLINE.equals(event.reason())
                        ? "driver went offline" : "offer expired";
                applyOrIgnore(event.rideId(), RideTransition.RETURN_TO_MATCHING, event.driverId(), reason);
            }
            case RIDE_ACCEPTED -> {
                RideAcceptedV1 event = codec.payload(envelope, RideAcceptedV1.class);
                applyOrIgnore(event.rideId(), RideTransition.ACCEPT, event.driverId(), "driver accepted");
            }
            default -> throw new IllegalArgumentException("Not a ride projection event: " + type);
        }
    }

    private void applyOrIgnore(UUID rideId, RideTransition transition, String driverId, String reason) {
        if (rides.apply(rideId, transition, driverId, reason).isEmpty()) {
            String current = rides.find(rideId).map(ride -> ride.status().name()).orElse("unknown");
            log.info("Ignored {} for a ride in status {}", transition, current);
        }
    }
}
