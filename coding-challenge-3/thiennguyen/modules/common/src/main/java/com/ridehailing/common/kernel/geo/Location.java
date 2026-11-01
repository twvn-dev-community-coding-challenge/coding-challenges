package com.ridehailing.common.kernel.geo;

import com.ridehailing.common.kernel.InvalidValueException;

public record Location(double latitude, double longitude) {

    public Location {
        if (!Double.isFinite(latitude) || latitude < -90.0 || latitude > 90.0) {
            throw new InvalidValueException("INVALID_LATITUDE", "latitude must be between -90 and 90");
        }
        if (!Double.isFinite(longitude) || longitude < -180.0 || longitude > 180.0) {
            throw new InvalidValueException("INVALID_LONGITUDE", "longitude must be between -180 and 180");
        }
    }
}
