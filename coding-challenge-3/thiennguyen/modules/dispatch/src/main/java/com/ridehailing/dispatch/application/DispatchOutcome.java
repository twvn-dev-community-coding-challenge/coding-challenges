package com.ridehailing.dispatch.application;

public enum DispatchOutcome {
    OFFERED,
    /** No eligible driver; the ride stays MATCHING with no assigned driver. */
    NO_DRIVER,
    /** The ride was not waiting for a driver (already offered, accepted, finished or unknown). */
    SKIPPED
}
