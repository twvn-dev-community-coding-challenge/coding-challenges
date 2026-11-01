package com.ridehailing.ride.domain;

import com.ridehailing.common.kernel.StateTransition;

import java.util.Collections;
import java.util.EnumSet;
import java.util.Set;

import static com.ridehailing.ride.domain.RideStatus.*;

/**
 * The ride lifecycle; every other status change is rejected. Each transition also says whether only the driver the ride
 * is assigned to may cause it, and what happens to {@code assignedDriverId}.
 */
public enum RideTransition implements StateTransition<RideStatus> {
    START_MATCHING(MATCHING, AssignedDriver.KEEP, false, REQUESTED),
    ASSIGN_DRIVER(DRIVER_ASSIGNED, AssignedDriver.SET, false, MATCHING),
    RETURN_TO_MATCHING(MATCHING, AssignedDriver.CLEAR, true, DRIVER_ASSIGNED),
    ACCEPT(DRIVER_ACCEPTED, AssignedDriver.KEEP, true, DRIVER_ASSIGNED),
    COMPLETE(COMPLETED, AssignedDriver.KEEP, false, DRIVER_ACCEPTED),
    CANCEL(CANCELLED, AssignedDriver.KEEP_IF_ACCEPTED, false, REQUESTED, MATCHING, DRIVER_ASSIGNED, DRIVER_ACCEPTED);

    /**
     * What a transition does to the ride's assigned driver.
     */
    public enum AssignedDriver {
        KEEP,
        /**
         * The driver the transition names becomes the assigned driver.
         */
        SET,
        CLEAR,
        /**
         * Kept only if the ride had been accepted (that driver drove towards the pickup); cleared otherwise.
         */
        KEEP_IF_ACCEPTED
    }

    private final RideStatus to;
    private final AssignedDriver assignedDriver;
    private final boolean requiresAssignedDriver;
    private final Set<RideStatus> from;

    RideTransition(RideStatus to, AssignedDriver assignedDriver, boolean requiresAssignedDriver, RideStatus first,
                   RideStatus... rest) {
        this.to = to;
        this.assignedDriver = assignedDriver;
        this.requiresAssignedDriver = requiresAssignedDriver;
        this.from = Collections.unmodifiableSet(EnumSet.of(first, rest));
    }

    @Override
    public RideStatus to() {
        return to;
    }

    @Override
    public Set<RideStatus> from() {
        return from;
    }

    public AssignedDriver assignedDriver() {
        return assignedDriver;
    }

    /**
     * Only the driver the ride is assigned to may cause this transition (a rejection or acceptance of its offer).
     */
    public boolean requiresAssignedDriver() {
        return requiresAssignedDriver;
    }

    /**
     * Whether the transition needs to know which driver it is about.
     */
    public boolean namesDriver() {
        return requiresAssignedDriver || assignedDriver == AssignedDriver.SET;
    }
}
