package com.ridehailing.ride.application;

import static org.assertj.core.api.Assertions.assertThat;

import com.ridehailing.common.kernel.geo.Location;
import org.junit.jupiter.api.Test;

class RequestFingerprintTest {

    private static final CreateRideCommand COMMAND =
            new CreateRideCommand("P001", new Location(10.7769, 106.7009), new Location(10.8231, 106.6297));

    @Test
    void identicalRequestsHaveTheSameFingerprint() {
        CreateRideCommand copy =
                new CreateRideCommand("P001", new Location(10.7769, 106.7009), new Location(10.8231, 106.6297));
        assertThat(RequestFingerprint.of(copy)).isEqualTo(RequestFingerprint.of(COMMAND));
    }

    @Test
    void anyDifferenceChangesTheFingerprint() {
        CreateRideCommand otherPassenger = new CreateRideCommand("P002", COMMAND.pickup(), COMMAND.destination());
        CreateRideCommand otherDestination =
                new CreateRideCommand("P001", COMMAND.pickup(), new Location(10.8231, 106.6298));
        assertThat(RequestFingerprint.of(otherPassenger)).isNotEqualTo(RequestFingerprint.of(COMMAND));
        assertThat(RequestFingerprint.of(otherDestination)).isNotEqualTo(RequestFingerprint.of(COMMAND));
    }

    @Test
    void isA64CharacterHexSha256() {
        assertThat(RequestFingerprint.of(COMMAND)).hasSize(64).matches("[0-9a-f]{64}");
    }
}
