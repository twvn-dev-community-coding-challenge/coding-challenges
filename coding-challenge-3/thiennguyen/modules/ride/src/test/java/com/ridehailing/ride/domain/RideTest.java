package com.ridehailing.ride.domain;

import static org.assertj.core.api.Assertions.assertThat;

import com.ridehailing.common.kernel.geo.Location;
import java.time.Instant;
import java.util.UUID;
import org.junit.jupiter.api.Test;

class RideTest {

    @Test
    void aRequestedRideWaitsForDispatchWithNoDriverAtTheFirstVersion() {
        UUID id = UUID.randomUUID();
        Location pickup = new Location(10.7769, 106.7009);
        Location destination = new Location(10.8231, 106.6297);
        Instant now = Instant.parse("2026-10-26T10:00:00Z");

        Ride ride = Ride.requested(id, "P001", pickup, destination, now);

        assertThat(ride).isEqualTo(
                new Ride(id, "P001", pickup, destination, RideStatus.REQUESTED, null, 0L, now, now));
    }
}
