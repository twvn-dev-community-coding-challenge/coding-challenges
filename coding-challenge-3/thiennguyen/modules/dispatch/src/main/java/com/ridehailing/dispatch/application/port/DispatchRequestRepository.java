package com.ridehailing.dispatch.application.port;

import com.ridehailing.common.kernel.geo.Location;
import com.ridehailing.dispatch.domain.DispatchRequest;
import com.ridehailing.dispatch.domain.DispatchRequestStatus;
import com.ridehailing.dispatch.domain.DispatchRequestTransition;
import java.time.Instant;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

/**
 * The dispatch aggregate's root store. Locking a request serialises every change to the ride's offers, exclusions and
 * held driver; it is always locked before the driver's supply row.
 */
public interface DispatchRequestRepository {

    /** Opens a WAITING request; {@code false} if dispatch already serves this ride, or it already ended. */
    boolean open(UUID rideId, Location pickup);

    /**
     * Records a ride that ended ({@code outcome} CANCELLED or COMPLETED) before dispatch opened its request, so a later
     * RideRequested is ignored. {@code false} if dispatch already has a request for this ride.
     */
    boolean recordEnded(UUID rideId, DispatchRequestStatus outcome);

    Optional<DispatchRequest> find(UUID rideId);

    /** Locks the request for the rest of the caller's transaction. */
    Optional<DispatchRequest> lock(UUID rideId);

    /**
     * Locks a request a matching attempt may serve (one {@link DispatchRequestTransition#OFFER} applies to) and marks
     * it as just tried, moving it out of the sweeper's reach. Empty when the request is not waiting for a driver.
     */
    Optional<DispatchRequest> lockForMatching(UUID rideId);

    /**
     * Applies {@code transition} if the request is in one of its {@code from} statuses and, when the transition
     * {@link DispatchRequestTransition#requiresHolder() requires the holder}, held by {@code driverId}.
     */
    boolean apply(UUID rideId, DispatchRequestTransition transition, String driverId);

    /** Requests still waiting for a driver that nobody has tried to match since {@code updatedBefore}. */
    List<UUID> findStaleWaiting(Instant updatedBefore, int limit);
}
