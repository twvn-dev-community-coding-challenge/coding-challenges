package com.ridehailing.dispatch.domain;

import java.util.UUID;

/**
 * A driver's assignment after a change in dispatch; {@code version} orders the published DriverAssignmentChanged.
 */
public record SupplyChange(String driverId, UUID rideId, Assignment assignment, long version) {
}
