package com.ridehailing.acceptance;

import static org.assertj.core.api.Assertions.assertThat;

import com.ridehailing.driver.domain.DriverStatus;
import com.ridehailing.ride.domain.RideHistoryEntry;
import com.ridehailing.ride.domain.RideStatus;
import com.ridehailing.support.Api;
import com.ridehailing.support.IntegrationTestSupport;
import com.ridehailing.support.TestCity;
import java.util.Arrays;
import java.util.UUID;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;

@Tag("AC-06")
class AC06StateMachineAcceptanceIT extends IntegrationTestSupport {

    private final TestCity city = TestCity.next();

    @AfterEach
    void invariantsHold() {
        assertInvariants();
    }

    @Test
    void rejectsCompletingARideThatWasNeverAccepted() {
        UUID rideId = createRide(fixtures.passenger(), city.centre()).id();
        awaitRideStatus(rideId, RideStatus.MATCHING);

        Api.Response response = api.post("/rides/" + rideId + "/complete");

        assertThat(response.status()).as(response.body()).isEqualTo(409);
        assertThat(response.code()).isEqualTo("INVALID_TRANSITION");
        assertThat(rideRow(rideId).status()).isEqualTo(RideStatus.MATCHING);
    }

    @Test
    void terminalStatesRejectFurtherTransitions() {
        String driver = fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(0.5));
        UUID rideId = createRide(fixtures.passenger(), city.centre()).id();
        awaitOfferedTo(rideId, driver);
        acceptAndAwait(rideId, driver);
        api.post("/rides/" + rideId + "/complete");
        eventually().untilAsserted(() -> assertThat(requestRow(rideId).status()).isEqualTo("COMPLETED"));

        assertThat(api.post("/rides/" + rideId + "/cancel").status()).isEqualTo(409);
        assertThat(accept(rideId, driver).status()).isEqualTo(409);
        assertThat(rideRow(rideId).status()).isEqualTo(RideStatus.COMPLETED);
    }

    @Test
    void duplicateTransitionsAreAnsweredWithoutChange() {
        String driver = fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(0.5));
        UUID rideId = createRide(fixtures.passenger(), city.centre()).id();
        awaitOfferedTo(rideId, driver);
        acceptAndAwait(rideId, driver);

        assertThat(api.post("/rides/" + rideId + "/complete").status()).isEqualTo(200);
        long versionAfterFirst = rideRow(rideId).version();
        assertThat(api.post("/rides/" + rideId + "/complete").status()).isEqualTo(200);

        assertThat(rideRow(rideId).version()).isEqualTo(versionAfterFirst);
        assertThat(outboxCount(rideId, "RideCompleted")).isEqualTo(1);
    }

    @Test
    void cancellingTwiceIsHarmless() {
        UUID rideId = createRide(fixtures.passenger(), city.centre()).id();

        assertThat(api.post("/rides/" + rideId + "/cancel").status()).isEqualTo(200);
        assertThat(api.post("/rides/" + rideId + "/cancel").status()).isEqualTo(200);
        assertThat(outboxCount(rideId, "RideCancelled")).isEqualTo(1);
    }

    @Test
    void persistsEveryStateChangeInOrder() {
        String first = fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(0.4));
        String second = fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(0.9));
        UUID rideId = createRide(fixtures.passenger(), city.centre()).id();
        awaitOfferedTo(rideId, first);
        reject(rideId, first);
        awaitOfferedTo(rideId, second);
        acceptAndAwait(rideId, second);
        api.post("/rides/" + rideId + "/complete");

        Api.Response response = api.get("/rides/" + rideId + "/history");

        assertThat(response.status()).as(response.body()).isEqualTo(200);
        RideHistoryEntry[] history = response.as(RideHistoryEntry[].class);
        assertThat(Arrays.stream(history).map(RideHistoryEntry::toStatus)).containsExactly(
                RideStatus.REQUESTED, RideStatus.MATCHING, RideStatus.DRIVER_ASSIGNED, RideStatus.MATCHING,
                RideStatus.DRIVER_ASSIGNED, RideStatus.DRIVER_ACCEPTED, RideStatus.COMPLETED);
    }

    @Test
    void returns404ForAnUnknownRide() {
        Api.Response response = api.post("/rides/" + UUID.randomUUID() + "/cancel");

        assertThat(response.status()).isEqualTo(404);
        assertThat(response.code()).isEqualTo("RIDE_NOT_FOUND");
    }
}
