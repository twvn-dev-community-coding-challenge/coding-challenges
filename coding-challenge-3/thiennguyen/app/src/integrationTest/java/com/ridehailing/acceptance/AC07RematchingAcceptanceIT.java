package com.ridehailing.acceptance;

import static org.assertj.core.api.Assertions.assertThat;

import com.ridehailing.driver.domain.DriverStatus;
import com.ridehailing.ride.domain.RideStatus;
import com.ridehailing.support.IntegrationTestSupport;
import com.ridehailing.support.TestCity;
import java.time.Duration;
import java.util.UUID;
import org.awaitility.Awaitility;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;

@Tag("AC-07")
class AC07RematchingAcceptanceIT extends IntegrationTestSupport {

    private final TestCity city = TestCity.next();

    @AfterEach
    void invariantsHold() {
        assertInvariants();
    }

    @Test
    void aRejectedRideIsOfferedToTheNextDriverWhoCanAccept() {
        String first = fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(0.5));
        String second = fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(1.0));
        UUID rideId = createRide(fixtures.passenger(), city.centre()).id();
        awaitOfferedTo(rideId, first);

        assertThat(reject(rideId, first).status()).isEqualTo(200);

        awaitOfferedTo(rideId, second);
        acceptAndAwait(rideId, second);
    }

    @Test
    void aDriverWhoRejectedIsNeverOfferedThatRideAgain() {
        String first = fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(0.5));
        String second = fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(1.0));
        UUID rideId = createRide(fixtures.passenger(), city.centre()).id();
        awaitOfferedTo(rideId, first);
        reject(rideId, first);
        awaitOfferedTo(rideId, second);

        reject(rideId, second);

        Awaitility.await().during(Duration.ofSeconds(2)).atMost(Duration.ofSeconds(10)).untilAsserted(() -> {
            assertThat(rideRow(rideId).status()).isEqualTo(RideStatus.MATCHING);
            assertThat(rideRow(rideId).assignedDriverId()).isNull();
            assertThat(offerRows(rideId)).containsExactly(
                    new OfferRow(first, "REJECTED"), new OfferRow(second, "REJECTED"));
        });
        awaitDriverStatus(first, DriverStatus.AVAILABLE);
    }

    @Test
    void aDriverWhoRejectedOneRideIsStillOfferedOthers() {
        String driver = fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(0.5));
        UUID firstRide = createRide(fixtures.passenger(), city.centre()).id();
        awaitOfferedTo(firstRide, driver);
        reject(firstRide, driver);
        awaitRideStatus(firstRide, RideStatus.MATCHING);

        UUID secondRide = createRide(fixtures.passenger(), city.centre()).id();

        awaitOfferedTo(secondRide, driver);
    }
}
