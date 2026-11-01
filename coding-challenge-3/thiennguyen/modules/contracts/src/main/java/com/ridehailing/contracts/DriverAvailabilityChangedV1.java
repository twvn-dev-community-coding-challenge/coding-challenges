package com.ridehailing.contracts;

/**
 * The driver went online or offline. {@code version} is the driver row version, increasing per driver.
 */
public record DriverAvailabilityChangedV1(String driverId, boolean online, long version) {
}
