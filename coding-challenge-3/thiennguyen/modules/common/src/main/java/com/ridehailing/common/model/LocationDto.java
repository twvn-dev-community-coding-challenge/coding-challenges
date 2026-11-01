package com.ridehailing.common.model;

import com.ridehailing.common.kernel.geo.Location;
import jakarta.validation.constraints.DecimalMax;
import jakarta.validation.constraints.DecimalMin;
import jakarta.validation.constraints.NotNull;

public record LocationDto(
        @NotNull @DecimalMin("-90.0") @DecimalMax("90.0") Double latitude,
        @NotNull @DecimalMin("-180.0") @DecimalMax("180.0") Double longitude) {

    public Location toLocation() {
        return new Location(latitude, longitude);
    }

    public static LocationDto from(Location location) {
        return location == null ? null : new LocationDto(location.latitude(), location.longitude());
    }
}
