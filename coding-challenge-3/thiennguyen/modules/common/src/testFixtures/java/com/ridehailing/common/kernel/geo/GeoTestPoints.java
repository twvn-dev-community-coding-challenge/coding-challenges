package com.ridehailing.common.kernel.geo;

/** Test helper: the point reached by travelling {@code distanceKm} from {@code origin} on {@code bearingDegrees}. */
public final class GeoTestPoints {

    private GeoTestPoints() {
    }

    public static Location destination(Location origin, double bearingDegrees, double distanceKm) {
        double angular = distanceKm / GeoDistance.EARTH_MEAN_RADIUS_KM;
        double bearing = Math.toRadians(bearingDegrees);
        double latitude = Math.toRadians(origin.latitude());
        double longitude = Math.toRadians(origin.longitude());
        double targetLatitude = Math.asin(Math.sin(latitude) * Math.cos(angular)
                + Math.cos(latitude) * Math.sin(angular) * Math.cos(bearing));
        double targetLongitude = longitude + Math.atan2(
                Math.sin(bearing) * Math.sin(angular) * Math.cos(latitude),
                Math.cos(angular) - Math.sin(latitude) * Math.sin(targetLatitude));
        double normalisedLongitude = ((Math.toDegrees(targetLongitude) + 540.0) % 360.0) - 180.0;
        return new Location(Math.toDegrees(targetLatitude), normalisedLongitude);
    }

    public static Location north(Location origin, double distanceKm) {
        return destination(origin, 0.0, distanceKm);
    }
}
