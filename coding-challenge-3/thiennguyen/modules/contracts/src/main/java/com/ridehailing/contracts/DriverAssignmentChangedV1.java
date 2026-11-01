package com.ridehailing.contracts;

import java.util.UUID;

/**
 * Dispatch's assignment of one driver changed. {@code version} increases with every change of that driver's
 * dispatch state, so consumers can drop out-of-order or duplicate updates. {@code rideId} is null for NONE.
 */
public record DriverAssignmentChangedV1(String driverId, UUID rideId, AssignmentState assignment, long version) {
}
