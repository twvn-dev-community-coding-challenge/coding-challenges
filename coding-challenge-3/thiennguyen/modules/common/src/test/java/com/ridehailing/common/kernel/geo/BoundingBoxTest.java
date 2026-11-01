package com.ridehailing.common.kernel.geo;

import static org.assertj.core.api.Assertions.assertThat;

import java.util.Random;
import org.junit.jupiter.api.Test;

class BoundingBoxTest {

    @Test
    void containsEveryPointWithinTheRadius() {
        Random random = new Random(42L);
        for (int sample = 0; sample < 20_000; sample++) {
            Location centre = new Location(random.nextDouble(-85, 85), random.nextDouble(-180, 180));
            double radiusKm = random.nextDouble(0.1, 50.0);
            Location point = GeoTestPoints.destination(
                    centre, random.nextDouble(0, 360), radiusKm * random.nextDouble(0, 0.999));
            BoundingBox box = BoundingBox.around(centre, radiusKm);
            assertThat(box.contains(point))
                    .as("box %s around %s (r=%.3f km) must contain %s", box, centre, radiusKm, point)
                    .isTrue();
        }
    }

    @Test
    void excludesAPointFarOutsideTheRadius() {
        Location centre = new Location(10.7769, 106.7009);
        assertThat(BoundingBox.around(centre, 2.0).contains(GeoTestPoints.north(centre, 5.0))).isFalse();
    }

    @Test
    void spansAllLongitudesNearAPole() {
        BoundingBox box = BoundingBox.around(new Location(89.99, 10.0), 5.0);
        assertThat(box.minLongitude()).isEqualTo(-180.0);
        assertThat(box.maxLongitude()).isEqualTo(180.0);
    }

    @Test
    void spansAllLongitudesAcrossTheAntimeridian() {
        BoundingBox box = BoundingBox.around(new Location(0.0, 179.99), 5.0);
        assertThat(box.minLongitude()).isEqualTo(-180.0);
        assertThat(box.maxLongitude()).isEqualTo(180.0);
    }
}
