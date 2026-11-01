package com.ridehailing.dispatch.domain;

import com.ridehailing.common.kernel.StateTransition;

import java.util.Set;

/**
 * A pending offer is answered exactly once: accepted, rejected, expired or cancelled with its ride.
 */
public enum OfferTransition implements StateTransition<OfferStatus> {
    ACCEPT(OfferStatus.ACCEPTED),
    REJECT(OfferStatus.REJECTED),
    EXPIRE(OfferStatus.EXPIRED),
    CANCEL(OfferStatus.CANCELLED);

    private static final Set<OfferStatus> PENDING_ONLY = Set.of(OfferStatus.PENDING);

    private final OfferStatus to;

    OfferTransition(OfferStatus to) {
        this.to = to;
    }

    @Override
    public OfferStatus to() {
        return to;
    }

    @Override
    public Set<OfferStatus> from() {
        return PENDING_ONLY;
    }
}
