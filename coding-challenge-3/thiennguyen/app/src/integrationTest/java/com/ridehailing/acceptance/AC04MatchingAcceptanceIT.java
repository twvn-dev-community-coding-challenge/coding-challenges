package com.ridehailing.acceptance;

import static org.assertj.core.api.Assertions.assertThat;

import com.ridehailing.driver.domain.DriverStatus;
import com.ridehailing.ride.domain.RideStatus;
import com.ridehailing.support.Fixtures;
import com.ridehailing.support.IntegrationTestSupport;
import com.ridehailing.support.TestCity;
import java.util.UUID;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;

@Tag("AC-04")
class AC04MatchingAcceptanceIT extends IntegrationTestSupport {

    private final TestCity city = TestCity.next();

    @AfterEach
    void invariantsHold() {
        assertInvariants();
    }

    @Test
    void offersTheRideToTheNearestDriverWithinTheRadius() {
        String driverA = fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(0.8));
        fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(1.5));
        fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(3.2));

        UUID rideId = createRide(fixtures.passenger(), city.centre()).id();

        awaitOfferedTo(rideId, driverA);
    }

    @Test
    void leavesTheRideMatchingWhenEveryDriverIsOutsideTheRadius() {
        String driverC = fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(3.2));

        UUID rideId = createRide(fixtures.passenger(), city.centre()).id();

        awaitRideStatus(rideId, RideStatus.MATCHING);
        assertThat(rideRow(rideId).assignedDriverId()).isNull();
        assertThat(offerRows(rideId)).isEmpty();
        assertThat(driverRow(driverC).status()).isEqualTo(DriverStatus.AVAILABLE);
    }

    @Test
    void breaksDistanceTiesByTheSmallestDriverId() {
        String prefix = "D-" + Fixtures.uniqueSuffix();
        fixtures.driver(prefix + "-b", DriverStatus.AVAILABLE, city.kmNorth(1.0));
        String smallest = fixtures.driver(prefix + "-a", DriverStatus.AVAILABLE, city.kmNorth(1.0));

        UUID rideId = createRide(fixtures.passenger(), city.centre()).id();

        awaitOfferedTo(rideId, smallest);
    }

    @Test
    void skipsDriversThatAreAlreadyBusy() {
        String nearest = fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(0.3));
        String next = fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(1.2));
        UUID firstRide = createRide(fixtures.passenger(), city.centre()).id();
        awaitOfferedTo(firstRide, nearest);
        assertThat(accept(firstRide, nearest).status()).isEqualTo(200);

        UUID secondRide = createRide(fixtures.passenger(), city.centre()).id();

        awaitOfferedTo(secondRide, next);
    }
}
