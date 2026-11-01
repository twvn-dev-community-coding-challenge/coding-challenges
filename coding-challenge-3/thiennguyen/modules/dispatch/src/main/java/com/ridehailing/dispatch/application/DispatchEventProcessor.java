package com.ridehailing.dispatch.application;

import com.ridehailing.common.events.EventCodec;
import com.ridehailing.common.events.ProcessedEvents;
import com.ridehailing.contracts.DriverAvailabilityChangedV1;
import com.ridehailing.contracts.DriverLocationUpdatedV1;
import com.ridehailing.contracts.EventEnvelope;
import com.ridehailing.contracts.EventType;
import com.ridehailing.contracts.OfferExpiredV1;
import com.ridehailing.contracts.OfferRejectedV1;
import com.ridehailing.contracts.RideCancelledV1;
import com.ridehailing.contracts.RideCompletedV1;
import com.ridehailing.contracts.RideRequestedV1;
import com.ridehailing.dispatch.domain.DispatchRequestTransition;
import java.util.EnumSet;
import java.util.Set;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;

/** De-duplicates on eventId and applies the event in the same transaction. */
@Component
@RequiredArgsConstructor
@Slf4j
public class DispatchEventProcessor {

    public static final String CONSUMER = "dispatch";
    public static final Set<EventType> HANDLED_EVENTS = EnumSet.of(
            EventType.RIDE_REQUESTED, EventType.RIDE_CANCELLED, EventType.RIDE_COMPLETED,
            EventType.OFFER_REJECTED, EventType.OFFER_EXPIRED,
            EventType.DRIVER_AVAILABILITY_CHANGED, EventType.DRIVER_LOCATION_UPDATED);

    private final ProcessedEvents processedEvents;
    private final EventCodec codec;
    private final DispatchService dispatch;
    private final RideOutcomes rideOutcomes;
    private final SupplyUpdates supplyUpdates;

    @Transactional
    public void process(EventType type, EventEnvelope envelope) {
        if (!processedEvents.markProcessed(CONSUMER, envelope.eventId())) {
            log.info("Duplicate delivery ignored; event was already processed");
            return;
        }
        switch (type) {
            case RIDE_REQUESTED -> dispatch.open(codec.payload(envelope, RideRequestedV1.class));
            case RIDE_CANCELLED -> rideOutcomes.rideEnded(
                    codec.payload(envelope, RideCancelledV1.class).rideId(), DispatchRequestTransition.CANCEL);
            case RIDE_COMPLETED -> rideOutcomes.rideEnded(
                    codec.payload(envelope, RideCompletedV1.class).rideId(), DispatchRequestTransition.COMPLETE);
            case OFFER_REJECTED -> dispatch.dispatch(codec.payload(envelope, OfferRejectedV1.class).rideId());
            case OFFER_EXPIRED -> dispatch.dispatch(codec.payload(envelope, OfferExpiredV1.class).rideId());
            case DRIVER_AVAILABILITY_CHANGED -> supplyUpdates.availabilityChanged(
                    codec.payload(envelope, DriverAvailabilityChangedV1.class));
            case DRIVER_LOCATION_UPDATED -> supplyUpdates.locationUpdated(
                    codec.payload(envelope, DriverLocationUpdatedV1.class));
            default -> throw new IllegalArgumentException("Dispatch does not handle " + type);
        }
    }
}
