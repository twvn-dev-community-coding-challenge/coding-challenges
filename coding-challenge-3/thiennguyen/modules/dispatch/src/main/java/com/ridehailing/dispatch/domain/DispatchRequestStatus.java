package com.ridehailing.dispatch.domain;

public enum DispatchRequestStatus {
    /**
     * Looking for a driver (the ride is MATCHING).
     */
    WAITING,
    /**
     * A driver holds a pending offer.
     */
    OFFERED,
    /**
     * The driver accepted; dispatch holds the driver BUSY until the ride ends.
     */
    ACCEPTED,
    CANCELLED,
    COMPLETED;

    public boolean isTerminal() {
        return this == CANCELLED || this == COMPLETED;
    }
}
