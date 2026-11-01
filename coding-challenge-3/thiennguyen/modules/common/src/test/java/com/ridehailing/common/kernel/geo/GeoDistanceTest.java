package com.ridehailing.common.kernel.geo;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.within;

import java.util.Random;
import org.junit.jupiter.api.Test;

class GeoDistanceTest {

    private static final Location PICKUP = new Location(10.7769, 106.7009);
    private static final Location DESTINATION = new Location(10.8231, 106.6297);

    @Test
    void distanceToSelfIsZero() {
        assertThat(GeoDistance.haversineKm(PICKUP, PICKUP)).isZero();
    }

    @Test
    void oneDegreeOfLatitudeIsAbout111Kilometres() {
        assertThat(GeoDistance.haversineKm(new Location(0, 0), new Location(1, 0))).isCloseTo(111.195, within(0.001));
    }

    @Test
    void matchesReferenceDistanceForTheBriefExample() {
        assertThat(GeoDistance.haversineKm(PICKUP, DESTINATION)).isCloseTo(9.3204, within(0.001));
    }

    @Test
    void travellingAKnownDistanceIsMeasuredBack() {
        assertThat(GeoDistance.haversineKm(PICKUP, GeoTestPoints.north(PICKUP, 0.8))).isCloseTo(0.8, within(1e-6));
    }

    @Test
    void isSymmetricAndSatisfiesTheTriangleInequality() {
        Random random = new Random(20261026L);
        for (int sample = 0; sample < 2_000; sample++) {
            Location a = randomLocation(random);
            Location b = randomLocation(random);
            Location c = randomLocation(random);
            double ab = GeoDistance.haversineKm(a, b);
            assertThat(ab).isCloseTo(GeoDistance.haversineKm(b, a), within(1e-9));
            assertThat(ab).isLessThanOrEqualTo(GeoDistance.haversineKm(a, c) + GeoDistance.haversineKm(c, b) + 1e-6);
        }
    }

    private static Location randomLocation(Random random) {
        return new Location(random.nextDouble(-90, 90), random.nextDouble(-180, 180));
    }
}
