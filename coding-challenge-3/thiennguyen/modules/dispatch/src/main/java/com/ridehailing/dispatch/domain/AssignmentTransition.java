package com.ridehailing.dispatch.domain;

import com.ridehailing.common.kernel.StateTransition;

import java.util.Collections;
import java.util.EnumSet;
import java.util.Set;

import static com.ridehailing.dispatch.domain.Assignment.*;

/**
 * How dispatch takes and gives back a driver. Claiming needs a free driver who is online and has a known location;
 * every other change needs the driver to be held for the ride in question.
 */
public enum AssignmentTransition implements StateTransition<Assignment> {
    CLAIM(OFFERED, NONE),
    START_TRIP(BUSY, OFFERED),
    RELEASE(NONE, OFFERED, BUSY);

    private final Assignment to;
    private final Set<Assignment> from;

    AssignmentTransition(Assignment to, Assignment first, Assignment... rest) {
        this.to = to;
        this.from = Collections.unmodifiableSet(EnumSet.of(first, rest));
    }

    @Override
    public Assignment to() {
        return to;
    }

    @Override
    public Set<Assignment> from() {
        return from;
    }

    /**
     * {@code true} for CLAIM: the driver must be free, online and located, and becomes held for the ride.
     */
    public boolean claimsFreeDriver() {
        return this == CLAIM;
    }
}
