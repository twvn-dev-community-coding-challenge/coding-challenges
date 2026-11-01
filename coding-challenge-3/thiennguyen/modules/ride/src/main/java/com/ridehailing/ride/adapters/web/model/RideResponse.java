package com.ridehailing.ride.adapters.web.model;

import com.ridehailing.common.model.LocationDto;
import com.ridehailing.ride.domain.Ride;
import com.ridehailing.ride.domain.RideStatus;

import java.time.Instant;
import java.util.UUID;

public record RideResponse(
        UUID id,
        String passengerId,
        LocationDto pickup,
        LocationDto destination,
        RideStatus status,
        String assignedDriverId,
        long version,
        Instant createdAt,
        Instant updatedAt) {

    public static RideResponse from(Ride ride) {
        return new RideResponse(ride.id(), ride.passengerId(), LocationDto.from(ride.pickup()),
                LocationDto.from(ride.destination()), ride.status(), ride.assignedDriverId(), ride.version(),
                ride.createdAt(), ride.updatedAt());
    }
}
