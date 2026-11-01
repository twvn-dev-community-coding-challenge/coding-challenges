package com.ridehailing.contracts;

import java.time.Duration;

public final class Topics {
    private Topics() {
    }

    public static final int PARTITIONS = 6;
    public static final int MAX_RETRIES = 3;
    public static final Duration RETENTION_DAYS = Duration.ofDays(7);


    /**
     * RideRequested, RideCancelled, RideCompleted. Key: rideId.
     */
    public static final String RIDE_EVENTS = "ride-events";

    /**
     * MatchingStarted, RideOffered, OfferRejected, OfferExpired, RideAccepted (key rideId) and
     * DriverAssignmentChanged (key driverId).
     */
    public static final String DISPATCH_EVENTS = "dispatch-events";

    /**
     * DriverAvailabilityChanged, DriverLocationUpdated. Key: driverId.
     */
    public static final String DRIVER_EVENTS = "driver-events";

    public static String deadLetter(String topic) {
        return topic + ".DLT";
    }
}
