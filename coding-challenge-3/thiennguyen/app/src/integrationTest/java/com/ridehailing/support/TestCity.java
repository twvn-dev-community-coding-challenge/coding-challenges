package com.ridehailing.support;

import com.ridehailing.common.kernel.geo.Location;
import java.util.concurrent.atomic.AtomicInteger;

/**
 * A private patch of the map for one test, at least 50 km from every other test's patch and from the seed data, so
 * matching never sees another test's drivers and tests need no clean-up.
 */
public final class TestCity {

    private static final AtomicInteger NEXT = new AtomicInteger();
    private static final double KM_PER_DEGREE = 111.195;

    private final Location centre;

    private TestCity(Location centre) {
        this.centre = centre;
    }

    public static TestCity next() {
        int index = NEXT.getAndIncrement();
        double latitude = -50.0 + (index / 100) * 0.5;
        double longitude = -170.0 + (index % 100) * 1.0;
        return new TestCity(new Location(latitude, longitude));
    }

    public Location centre() {
        return centre;
    }

    public Location kmNorth(double kilometres) {
        return new Location(centre.latitude() + kilometres / KM_PER_DEGREE, centre.longitude());
    }
}
