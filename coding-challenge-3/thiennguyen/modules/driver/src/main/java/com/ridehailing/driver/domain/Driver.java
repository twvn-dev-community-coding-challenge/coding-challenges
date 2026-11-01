package com.ridehailing.driver.domain;

import com.ridehailing.common.kernel.geo.Location;

import java.time.Instant;

/**
 * {@code availability} is what the driver chose (OFFLINE or AVAILABLE); {@code assignment} is dispatch's view of the
 * driver (none, a pending offer or a trip). {@code location} is null until the driver reports one.
 */
public record Driver(
        String id,
        DriverStatus availability,
        Location location,
        DriverAssignment assignment,
        long version,
        Instant updatedAt) {

    /**
     * The status of the brief: OFFERED or BUSY while dispatch holds the driver, otherwise the driver's choice.
     */
    public DriverStatus status() {
        if (assignment.state() == Assignment.OFFERED) {
            return DriverStatus.OFFERED;
        }
        if (assignment.state() == Assignment.BUSY) {
            return DriverStatus.BUSY;
        }
        return availability;
    }
}
