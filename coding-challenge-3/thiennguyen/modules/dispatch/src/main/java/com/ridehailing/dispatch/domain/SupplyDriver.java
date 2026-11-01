package com.ridehailing.dispatch.domain;

import com.ridehailing.common.kernel.geo.Location;
import com.ridehailing.dispatch.domain.matching.DriverCandidate;

import java.util.UUID;

public record SupplyDriver(
        String driverId,
        boolean online,
        Location location,
        long availabilityVersion,
        Assignment assignment,
        UUID currentRideId,
        long version) {

    public DriverCandidate toCandidate() {
        return new DriverCandidate(driverId, location, online, assignment);
    }
}
