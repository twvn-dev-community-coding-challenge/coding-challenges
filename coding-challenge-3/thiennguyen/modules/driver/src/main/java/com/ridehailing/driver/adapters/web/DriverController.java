package com.ridehailing.driver.adapters.web;

import com.ridehailing.common.model.LocationDto;
import com.ridehailing.driver.adapters.web.model.AvailabilityRequest;
import com.ridehailing.driver.adapters.web.model.DriverResponse;
import com.ridehailing.driver.application.DriverService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/drivers")
@RequiredArgsConstructor
public class DriverController {

    private final DriverService drivers;

    @GetMapping("/{driverId}")
    public DriverResponse get(@PathVariable String driverId) {
        return DriverResponse.from(drivers.get(driverId));
    }

    @PutMapping("/{driverId}/availability")
    public DriverResponse changeAvailability(
            @PathVariable String driverId, @Valid @RequestBody AvailabilityRequest request) {
        return DriverResponse.from(drivers.changeAvailability(driverId, request.status()));
    }

    @PutMapping("/{driverId}/location")
    public DriverResponse updateLocation(@PathVariable String driverId, @Valid @RequestBody LocationDto request) {
        return DriverResponse.from(drivers.updateLocation(driverId, request.toLocation()));
    }
}
