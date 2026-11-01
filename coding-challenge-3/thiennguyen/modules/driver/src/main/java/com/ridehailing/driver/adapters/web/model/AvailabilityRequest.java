package com.ridehailing.driver.adapters.web.model;

import com.ridehailing.driver.domain.DriverStatus;
import jakarta.validation.constraints.NotNull;

public record AvailabilityRequest(@NotNull DriverStatus status) {
}
