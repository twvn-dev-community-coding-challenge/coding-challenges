package com.ridehailing.ride.domain;

import java.time.Instant;

public record RideHistoryEntry(
        RideStatus fromStatus, RideStatus toStatus, String driverId, String reason, Instant occurredAt) {
}
