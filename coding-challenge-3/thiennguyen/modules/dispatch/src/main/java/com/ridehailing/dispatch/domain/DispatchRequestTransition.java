package com.ridehailing.dispatch.domain;

import com.ridehailing.common.kernel.StateTransition;

import java.util.Collections;
import java.util.EnumSet;
import java.util.Set;

import static com.ridehailing.dispatch.domain.DispatchRequestStatus.*;

/**
 * The lifecycle of a dispatch request: offered to one driver at a time, accepted, or back to waiting after a rejection
 * or expiry, until the ride ends. Each transition says whether only the driver holding the offer may cause it and what
 * happens to the request's driver.
 */
public enum DispatchRequestTransition implements StateTransition<DispatchRequestStatus> {
    OFFER(OFFERED, HeldDriver.SET, false, WAITING),
    ACCEPT(ACCEPTED, HeldDriver.KEEP, true, OFFERED),
    RETURN_TO_WAITING(WAITING, HeldDriver.CLEAR, true, OFFERED),
    /**
     * The ride was cancelled; the last driver is kept for history.
     */
    CANCEL(CANCELLED, HeldDriver.KEEP, false, WAITING, OFFERED, ACCEPTED),
    /**
     * The trip was completed; the last driver is kept for history.
     */
    COMPLETE(COMPLETED, HeldDriver.KEEP, false, WAITING, OFFERED, ACCEPTED);

    /**
     * What a transition does to the driver the request holds.
     */
    public enum HeldDriver {
        KEEP,
        /**
         * The driver the transition names becomes the one holding the offer.
         */
        SET,
        CLEAR
    }

    private final DispatchRequestStatus to;
    private final HeldDriver heldDriver;
    private final boolean requiresHolder;
    private final Set<DispatchRequestStatus> from;

    DispatchRequestTransition(DispatchRequestStatus to, HeldDriver heldDriver, boolean requiresHolder,
                              DispatchRequestStatus first, DispatchRequestStatus... rest) {
        this.to = to;
        this.heldDriver = heldDriver;
        this.requiresHolder = requiresHolder;
        this.from = Collections.unmodifiableSet(EnumSet.of(first, rest));
    }

    @Override
    public DispatchRequestStatus to() {
        return to;
    }

    @Override
    public Set<DispatchRequestStatus> from() {
        return from;
    }

    public HeldDriver heldDriver() {
        return heldDriver;
    }

    /**
     * Only the driver holding the current offer may cause this transition (an acceptance, rejection or expiry).
     */
    public boolean requiresHolder() {
        return requiresHolder;
    }
}
