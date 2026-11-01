package com.ridehailing.dispatch.domain.matching;

import com.ridehailing.common.kernel.geo.Location;
import com.ridehailing.dispatch.domain.Assignment;

/**
 * Dispatch's own view of a driver it might offer a ride to; {@code location} is null until one was reported.
 */
public record DriverCandidate(String driverId, Location location, boolean online, Assignment assignment) {

    public boolean isFree() {
        return online && assignment == Assignment.NONE && location != null;
    }
}
