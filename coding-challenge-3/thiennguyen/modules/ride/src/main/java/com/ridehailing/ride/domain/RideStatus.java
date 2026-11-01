package com.ridehailing.ride.domain;

public enum RideStatus {
    REQUESTED,
    MATCHING,
    DRIVER_ASSIGNED,
    DRIVER_ACCEPTED,
    CANCELLED,
    COMPLETED;

    public boolean isTerminal() {
        return this == CANCELLED || this == COMPLETED;
    }
}
