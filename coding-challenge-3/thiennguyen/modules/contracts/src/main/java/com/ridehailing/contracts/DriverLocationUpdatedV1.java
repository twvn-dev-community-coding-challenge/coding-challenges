package com.ridehailing.contracts;

/**
 * The driver reported a new location. {@code version} is the driver row version, increasing per driver.
 */
public record DriverLocationUpdatedV1(String driverId, GeoPoint location, long version) {
}
