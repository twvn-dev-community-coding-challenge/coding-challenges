package com.ridehailing.dispatch.domain.matching;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.assertj.core.api.Assertions.within;

import com.ridehailing.common.kernel.geo.GeoDistance;
import com.ridehailing.common.kernel.geo.GeoTestPoints;
import com.ridehailing.common.kernel.geo.Location;
import com.ridehailing.dispatch.domain.Assignment;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.List;
import java.util.Random;
import java.util.Set;
import org.junit.jupiter.api.Test;

class NearestDriverMatcherTest {

    private static final Location PICKUP = new Location(10.7769, 106.7009);
    private final NearestDriverMatcher matcher = new NearestDriverMatcher(2.0);

    @Test
    void selectsTheNearestDriverAndExcludesThoseOutsideTheRadius() {
        List<DriverCandidate> drivers = List.of(free("B", 1.5), free("C", 3.2), free("A", 0.8));

        List<RankedCandidate> ranked = matcher.rank(PICKUP, drivers, Set.of());

        assertThat(ranked).extracting(RankedCandidate::driverId).containsExactly("A", "B");
        assertThat(ranked.getFirst().distanceKm()).isCloseTo(0.8, within(1e-6));
    }

    @Test
    void excludesOfflineDrivers() {
        DriverCandidate offline = new DriverCandidate("X", GeoTestPoints.north(PICKUP, 0.1), false, Assignment.NONE);
        assertThat(matcher.rank(PICKUP, List.of(offline), Set.of())).isEmpty();
    }

    @Test
    void excludesBusyAndOfferedDrivers() {
        List<DriverCandidate> drivers = List.of(
                new DriverCandidate("BUSY", GeoTestPoints.north(PICKUP, 0.1), true, Assignment.BUSY),
                new DriverCandidate("OFFER", GeoTestPoints.north(PICKUP, 0.2), true, Assignment.OFFERED));
        assertThat(matcher.rank(PICKUP, drivers, Set.of())).isEmpty();
    }

    @Test
    void selectsNobodyWhenNoSuitableDriverExists() {
        assertThat(matcher.rank(PICKUP, List.of(), Set.of())).isEmpty();
        assertThat(matcher.rank(PICKUP, List.of(free("FAR", 2.5)), Set.of())).isEmpty();
    }

    @Test
    void breaksDistanceTiesByTheSmallestDriverId() {
        Location sameSpot = GeoTestPoints.north(PICKUP, 1.0);
        List<DriverCandidate> drivers = List.of(at("D-b", sameSpot), at("D-c", sameSpot), at("D-a", sameSpot));
        assertThat(matcher.rank(PICKUP, drivers, Set.of())).extracting(RankedCandidate::driverId)
                .containsExactly("D-a", "D-b", "D-c");
    }

    @Test
    void skipsDriversExcludedForThisRide() {
        List<DriverCandidate> drivers = List.of(free("A", 0.5), free("B", 1.0));
        assertThat(matcher.rank(PICKUP, drivers, Set.of("A"))).extracting(RankedCandidate::driverId)
                .containsExactly("B");
    }

    @Test
    void skipsDriversWithoutAKnownLocation() {
        DriverCandidate noLocation = new DriverCandidate("NOLOC", null, true, Assignment.NONE);
        assertThat(matcher.rank(PICKUP, List.of(noLocation), Set.of())).isEmpty();
    }

    @Test
    void includesADriverExactlyOnTheRadius() {
        DriverCandidate edge = at("EDGE", GeoTestPoints.north(PICKUP, 1.999999));
        assertThat(matcher.rank(PICKUP, List.of(edge), Set.of())).hasSize(1);
    }

    @Test
    void rejectsANonPositiveRadius() {
        assertThatThrownBy(() -> new NearestDriverMatcher(0)).isInstanceOf(IllegalArgumentException.class);
    }

    @Test
    void rankingIsSortedEligibleAndStartsWithTheTrueNearest() {
        Random random = new Random(99L);
        Assignment[] assignments = Assignment.values();
        for (int run = 0; run < 500; run++) {
            List<DriverCandidate> drivers = new ArrayList<>();
            for (int index = 0; index < 30; index++) {
                Location location = GeoTestPoints.destination(PICKUP, random.nextDouble(0, 360), random.nextDouble(0, 4));
                drivers.add(new DriverCandidate("D" + random.nextInt(1000), location, random.nextBoolean(),
                        assignments[random.nextInt(assignments.length)]));
            }
            Set<String> excluded = Set.of("D" + random.nextInt(1000));

            List<RankedCandidate> ranked = matcher.rank(PICKUP, drivers, excluded);

            assertThat(ranked).isSortedAccordingTo(
                    Comparator.comparingDouble(RankedCandidate::distanceKm).thenComparing(RankedCandidate::driverId));
            assertThat(ranked).allSatisfy(candidate -> {
                assertThat(candidate.distanceKm()).isLessThanOrEqualTo(2.0);
                assertThat(excluded).doesNotContain(candidate.driverId());
            });
            drivers.stream()
                    .filter(driver -> driver.isFree() && !excluded.contains(driver.driverId()))
                    .mapToDouble(driver -> GeoDistance.haversineKm(PICKUP, driver.location()))
                    .filter(distance -> distance <= 2.0)
                    .min()
                    .ifPresent(nearest -> assertThat(ranked.getFirst().distanceKm()).isEqualTo(nearest));
        }
    }

    private static DriverCandidate free(String id, double kmNorth) {
        return at(id, GeoTestPoints.north(PICKUP, kmNorth));
    }

    private static DriverCandidate at(String id, Location location) {
        return new DriverCandidate(id, location, true, Assignment.NONE);
    }
}
