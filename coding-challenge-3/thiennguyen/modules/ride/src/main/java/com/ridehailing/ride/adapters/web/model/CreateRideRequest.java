package com.ridehailing.ride.adapters.web.model;

import com.ridehailing.common.model.LocationDto;
import com.ridehailing.ride.application.CreateRideCommand;
import jakarta.validation.Valid;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Size;

public record CreateRideRequest(
        @NotBlank @Size(max = 64) String passengerId,
        @NotNull @Valid LocationDto pickup,
        @NotNull @Valid LocationDto destination) {

    public CreateRideCommand toCommand() {
        return new CreateRideCommand(passengerId, pickup.toLocation(), destination.toLocation());
    }
}
