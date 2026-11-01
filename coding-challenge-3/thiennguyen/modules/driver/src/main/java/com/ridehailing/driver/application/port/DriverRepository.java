package com.ridehailing.driver.application.port;

import com.ridehailing.common.kernel.geo.Location;
import com.ridehailing.driver.domain.Driver;
import com.ridehailing.driver.domain.DriverAssignment;
import com.ridehailing.driver.domain.DriverTransition;

import java.util.Optional;

/**
 * The driver aggregate's store: availability and location are owned here, the assignment is dispatch's.
 */
public interface DriverRepository {

    Optional<Driver> find(String driverId);

    /**
     * Locks the driver for the rest of the caller's transaction.
     */
    Optional<Driver> findForUpdate(String driverId);

    /**
     * OFFLINE ⇄ AVAILABLE compare-and-set on {@code transition.from()}; going offline also requires that the driver's
     * copy of the assignment is {@code NONE}. Empty when the driver was not in that state.
     */
    Optional<Driver> applyAvailability(String driverId, DriverTransition transition);

    Optional<Driver> updateLocation(String driverId, Location location);

    /**
     * Applies dispatch's assignment if {@code assignmentVersion} is newer than the copy held; empty otherwise.
     */
    Optional<Driver> applyAssignment(String driverId, DriverAssignment assignment, long assignmentVersion);
}
