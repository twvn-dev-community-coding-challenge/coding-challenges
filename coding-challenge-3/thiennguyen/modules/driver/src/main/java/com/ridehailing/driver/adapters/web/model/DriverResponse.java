package com.ridehailing.driver.adapters.web.model;

import com.ridehailing.common.model.LocationDto;
import com.ridehailing.driver.domain.Driver;
import com.ridehailing.driver.domain.DriverStatus;
import java.time.Instant;
import java.util.UUID;

public record DriverResponse(
        String id, DriverStatus status, LocationDto location, UUID currentRideId, long version, Instant updatedAt) {

    public static DriverResponse from(Driver driver) {
        return new DriverResponse(driver.id(), driver.status(), LocationDto.from(driver.location()),
                driver.assignment().rideId(), driver.version(), driver.updatedAt());
    }
}
