package com.ridehailing.driver.application.port;

import com.ridehailing.driver.domain.Driver;

import java.util.Optional;

/**
 * Read model for polling clients; a newer version always wins and the cache never decides a change.
 */
public interface DriverReadCache {

    Optional<Driver> get(String driverId);

    void put(Driver driver);

    /**
     * Stores the driver once the surrounding transaction commits; a rolled-back change never reaches the cache.
     */
    void putAfterCommit(Driver driver);
}
