package com.ridehailing.dispatch.domain.matching;

import com.ridehailing.common.kernel.geo.GeoDistance;
import com.ridehailing.common.kernel.geo.Location;

import java.util.Collection;
import java.util.Comparator;
import java.util.List;
import java.util.Set;

/**
 * Nearest available driver within a radius: only drivers who are online, not holding an offer or a trip, with a known
 * location and not excluded for this ride, at most {@code maxRadiusKm} away (inclusive), ordered by Haversine distance
 * and then by the lexicographically smallest driver id.
 */
public final class NearestDriverMatcher {

    private static final Comparator<RankedCandidate> NEAREST_FIRST =
            Comparator.comparingDouble(RankedCandidate::distanceKm).thenComparing(RankedCandidate::driverId);

    private final double maxRadiusKm;

    public NearestDriverMatcher(double maxRadiusKm) {
        if (!(maxRadiusKm > 0)) {
            throw new IllegalArgumentException("maxRadiusKm must be positive");
        }
        this.maxRadiusKm = maxRadiusKm;
    }

    public double maxRadiusKm() {
        return maxRadiusKm;
    }

    public List<RankedCandidate> rank(
            Location pickup, Collection<DriverCandidate> candidates, Set<String> excludedDriverIds) {
        return candidates.stream()
                .filter(DriverCandidate::isFree)
                .filter(candidate -> !excludedDriverIds.contains(candidate.driverId()))
                .map(candidate -> new RankedCandidate(
                        candidate.driverId(), GeoDistance.haversineKm(pickup, candidate.location())))
                .filter(candidate -> candidate.distanceKm() <= maxRadiusKm)
                .sorted(NEAREST_FIRST)
                .toList();
    }
}
