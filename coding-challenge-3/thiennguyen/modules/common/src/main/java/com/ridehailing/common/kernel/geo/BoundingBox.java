package com.ridehailing.common.kernel.geo;

/**
 * Smallest latitude/longitude rectangle containing every point within a radius of a centre, used as an indexed SQL
 * pre-filter before exact Haversine ranking. Falls back to the full longitude range near the poles and across the
 * antimeridian, where a single rectangle cannot describe the circle.
 */
public record BoundingBox(double minLatitude, double maxLatitude, double minLongitude, double maxLongitude) {

    public static BoundingBox around(Location centre, double radiusKm) {
        double angularRadius = radiusKm / GeoDistance.EARTH_MEAN_RADIUS_KM;
        double latitude = Math.toRadians(centre.latitude());
        double minLatitude = latitude - angularRadius;
        double maxLatitude = latitude + angularRadius;
        if (minLatitude <= -Math.PI / 2 || maxLatitude >= Math.PI / 2) {
            return new BoundingBox(
                    Math.max(-90.0, Math.toDegrees(minLatitude)), Math.min(90.0, Math.toDegrees(maxLatitude)),
                    -180.0, 180.0);
        }
        double longitudeSpread = Math.sin(angularRadius) / Math.cos(latitude);
        if (longitudeSpread >= 1.0) {
            return new BoundingBox(Math.toDegrees(minLatitude), Math.toDegrees(maxLatitude), -180.0, 180.0);
        }
        double deltaLongitude = Math.toDegrees(Math.asin(longitudeSpread));
        double minLongitude = centre.longitude() - deltaLongitude;
        double maxLongitude = centre.longitude() + deltaLongitude;
        if (minLongitude < -180.0 || maxLongitude > 180.0) {
            return new BoundingBox(Math.toDegrees(minLatitude), Math.toDegrees(maxLatitude), -180.0, 180.0);
        }
        return new BoundingBox(Math.toDegrees(minLatitude), Math.toDegrees(maxLatitude), minLongitude, maxLongitude);
    }

    public boolean contains(Location location) {
        return location.latitude() >= minLatitude && location.latitude() <= maxLatitude
                && location.longitude() >= minLongitude && location.longitude() <= maxLongitude;
    }
}
