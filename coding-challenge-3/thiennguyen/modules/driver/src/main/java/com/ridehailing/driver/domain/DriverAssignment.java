package com.ridehailing.driver.domain;

import java.util.UUID;

/**
 * The driver's assignment as last published by dispatch: dispatch decides it, the driver module keeps a versioned copy
 * so its API can show OFFERED/BUSY and refuse to go offline mid-offer or mid-trip. {@code rideId} is the ride dispatch
 * holds the driver for, null when {@code state} is {@code NONE}.
 */
public record DriverAssignment(Assignment state, UUID rideId) {

    public static final DriverAssignment NONE = new DriverAssignment(Assignment.NONE, null);

    public static DriverAssignment of(Assignment state, UUID rideId) {
        return state == Assignment.NONE ? NONE : new DriverAssignment(state, rideId);
    }
}
