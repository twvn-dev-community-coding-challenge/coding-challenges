package com.ridehailing.contracts;

import java.util.Arrays;
import java.util.Optional;

/**
 * Every event in the system: its wire name, the topic its owner publishes it to, and its payload schema.
 */
public enum EventType {
    RIDE_REQUESTED("RideRequested", Topics.RIDE_EVENTS, RideRequestedV1.class),
    RIDE_CANCELLED("RideCancelled", Topics.RIDE_EVENTS, RideCancelledV1.class),
    RIDE_COMPLETED("RideCompleted", Topics.RIDE_EVENTS, RideCompletedV1.class),

    MATCHING_STARTED("MatchingStarted", Topics.DISPATCH_EVENTS, MatchingStartedV1.class),
    RIDE_OFFERED("RideOffered", Topics.DISPATCH_EVENTS, RideOfferedV1.class),
    OFFER_REJECTED("OfferRejected", Topics.DISPATCH_EVENTS, OfferRejectedV1.class),
    OFFER_EXPIRED("OfferExpired", Topics.DISPATCH_EVENTS, OfferExpiredV1.class),
    RIDE_ACCEPTED("RideAccepted", Topics.DISPATCH_EVENTS, RideAcceptedV1.class),
    DRIVER_ASSIGNMENT_CHANGED("DriverAssignmentChanged", Topics.DISPATCH_EVENTS, DriverAssignmentChangedV1.class),

    DRIVER_AVAILABILITY_CHANGED("DriverAvailabilityChanged", Topics.DRIVER_EVENTS, DriverAvailabilityChangedV1.class),
    DRIVER_LOCATION_UPDATED("DriverLocationUpdated", Topics.DRIVER_EVENTS, DriverLocationUpdatedV1.class);

    private final String wireName;
    private final String topic;
    private final Class<?> payloadType;

    EventType(String wireName, String topic, Class<?> payloadType) {
        this.wireName = wireName;
        this.topic = topic;
        this.payloadType = payloadType;
    }

    public String wireName() {
        return wireName;
    }

    public String topic() {
        return topic;
    }

    public Class<?> payloadType() {
        return payloadType;
    }

    public static Optional<EventType> fromWireName(String wireName) {
        return Arrays.stream(values()).filter(type -> type.wireName.equals(wireName)).findFirst();
    }
}
