package com.ridehailing.dispatch.application.port;

import com.ridehailing.common.kernel.geo.BoundingBox;
import com.ridehailing.common.kernel.geo.Location;
import com.ridehailing.dispatch.domain.AssignmentTransition;
import com.ridehailing.dispatch.domain.SupplyChange;
import com.ridehailing.dispatch.domain.SupplyDriver;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

/**
 * Dispatch's supply: each driver's online flag and location, copied from driver-events and guarded by their own
 * versions, and the assignment, owned here and changed only by compare-and-set.
 */
public interface DriverSupplyRepository {

    Optional<SupplyDriver> find(String driverId);

    /** Free drivers (online, no offer or trip, known location) inside the box: the pre-filter before ranking. */
    List<SupplyDriver> findFreeWithin(BoundingBox box);

    /**
     * Applies {@code transition} to the driver's assignment for {@code rideId}. Empty if the driver is no longer in a
     * {@code from} state for that ride: claimed by another dispatch, offline, or released already.
     */
    Optional<SupplyChange> apply(String driverId, UUID rideId, AssignmentTransition transition);

    /** Applies DriverAvailabilityChanged unless a newer availability was already applied. */
    void applyAvailability(String driverId, boolean online, long availabilityVersion);

    /** Applies DriverLocationUpdated unless a newer location was already applied. */
    void applyLocation(String driverId, Location location, long locationVersion);
}
