package com.ridehailing.common.kernel.geo;

public final class GeoDistance {

    public static final double EARTH_MEAN_RADIUS_KM = 6371.0088;

    private GeoDistance() {
    }

    public static double haversineKm(Location from, Location to) {
        double fromLatitude = Math.toRadians(from.latitude());
        double toLatitude = Math.toRadians(to.latitude());
        double deltaLatitude = toLatitude - fromLatitude;
        double deltaLongitude = Math.toRadians(to.longitude() - from.longitude());
        double haversine = Math.pow(Math.sin(deltaLatitude / 2), 2)
                + Math.cos(fromLatitude) * Math.cos(toLatitude) * Math.pow(Math.sin(deltaLongitude / 2), 2);
        return 2 * EARTH_MEAN_RADIUS_KM * Math.asin(Math.min(1.0, Math.sqrt(haversine)));
    }
}
