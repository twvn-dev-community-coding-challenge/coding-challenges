package com.ridehailing.dispatch.domain;

import com.ridehailing.common.kernel.TransitionOutcome;
import com.ridehailing.common.kernel.geo.Location;

import java.time.Instant;
import java.util.UUID;

/**
 * Dispatch's own record of a ride it serves, and the root of the dispatch aggregate: its offers and exclusions change
 * only while this request is locked. {@code driverId} holds the offered or accepted driver. {@code pickup} is null only
 * for a ride that ended before dispatch opened its request.
 */
public record DispatchRequest(
        UUID rideId,
        Location pickup,
        DispatchRequestStatus status,
        String driverId,
        long version,
        Instant createdAt,
        Instant updatedAt) {

    public boolean isHeldBy(String candidateDriverId) {
        return driverId != null && driverId.equals(candidateDriverId);
    }

    /**
     * Whether {@code transition}, caused by {@code actingDriverId} (null when no driver is acting), applies now, was
     * already applied, or is not allowed. A transition reserved to the offer's holder is rejected for anyone else.
     */
    public TransitionOutcome evaluate(DispatchRequestTransition transition, String actingDriverId) {
        if (transition.requiresHolder() && !isHeldBy(actingDriverId)) {
            return TransitionOutcome.REJECTED;
        }
        return transition.evaluate(status);
    }
}
