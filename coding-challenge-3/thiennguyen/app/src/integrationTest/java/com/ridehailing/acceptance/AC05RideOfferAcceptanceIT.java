package com.ridehailing.acceptance;

import static org.assertj.core.api.Assertions.assertThat;

import com.ridehailing.driver.domain.DriverStatus;
import com.ridehailing.ride.adapters.web.model.RideResponse;
import com.ridehailing.ride.domain.RideStatus;
import com.ridehailing.support.Api;
import com.ridehailing.support.IntegrationTestSupport;
import com.ridehailing.support.TestCity;
import java.time.Duration;
import java.util.Map;
import java.util.UUID;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;

@Tag("AC-05")
class AC05RideOfferAcceptanceIT extends IntegrationTestSupport {

    private final TestCity city = TestCity.next();
    private String driver;
    private UUID rideId;

    @BeforeEach
    void offerARide() {
        driver = fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(0.5));
        rideId = createRide(fixtures.passenger(), city.centre()).id();
        awaitOfferedTo(rideId, driver);
    }

    @AfterEach
    void invariantsHold() {
        assertInvariants();
    }

    @Test
    void creatingAnOfferAssignsTheRideAndMarksTheDriverOffered() {
        assertThat(rideRow(rideId).assignedDriverId()).isEqualTo(driver);
        assertThat(requestRow(rideId)).isEqualTo(new RequestRow("OFFERED", driver));
        assertThat(offerRows(rideId)).containsExactly(new OfferRow(driver, "PENDING"));
        awaitDriverStatus(driver, DriverStatus.OFFERED);
        assertThat(driverRow(driver).currentRideId()).isEqualTo(rideId);
    }

    @Test
    void onlyTheDriverHoldingTheOfferCanRespond() {
        String other = fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(1.9));

        Api.Response acceptByOther = accept(rideId, other);
        Api.Response rejectByOther = reject(rideId, other);

        assertThat(acceptByOther.status()).as(acceptByOther.body()).isEqualTo(409);
        assertThat(acceptByOther.code()).isEqualTo("OFFER_NOT_HELD");
        assertThat(rejectByOther.status()).isEqualTo(409);
        assertThat(requestRow(rideId).driverId()).isEqualTo(driver);
    }

    @Test
    void acceptingMakesTheRideAcceptedAndTheDriverBusy() {
        Api.Response response = accept(rideId, driver);

        assertThat(response.status()).as(response.body()).isEqualTo(200);
        assertThat(offerStatus(response)).isEqualTo("ACCEPTED");
        assertThat(offerRows(rideId)).containsExactly(new OfferRow(driver, "ACCEPTED"));
        awaitRideStatus(rideId, RideStatus.DRIVER_ACCEPTED);
        awaitDriverStatus(driver, DriverStatus.BUSY);
    }

    @Test
    void rejectingReturnsTheRideToMatchingAndTheDriverToAvailable() {
        Api.Response response = reject(rideId, driver);

        assertThat(response.status()).as(response.body()).isEqualTo(200);
        assertThat(offerStatus(response)).isEqualTo("REJECTED");
        assertThat(offerRows(rideId)).containsExactly(new OfferRow(driver, "REJECTED"));
        awaitRideStatus(rideId, RideStatus.MATCHING);
        assertThat(rideRow(rideId).assignedDriverId()).isNull();
        awaitDriverStatus(driver, DriverStatus.AVAILABLE);
    }

    @Test
    void anUnansweredOfferExpiresLikeARejection() {
        clock.advance(Duration.ofSeconds(31));

        eventually().untilAsserted(() -> {
            assertThat(offerRows(rideId)).containsExactly(new OfferRow(driver, "EXPIRED"));
            assertThat(rideRow(rideId).status()).isEqualTo(RideStatus.MATCHING);
            assertThat(rideRow(rideId).assignedDriverId()).isNull();
            assertThat(driverRow(driver).status()).isEqualTo(DriverStatus.AVAILABLE);
        });
        Api.Response lateAccept = accept(rideId, driver);
        assertThat(lateAccept.status()).as(lateAccept.body()).isEqualTo(409);
    }

    @Test
    void aDriverWhoGoesOfflineWhileOfferedHasTheOfferExpired() {
        awaitDriverStatus(driver, DriverStatus.OFFERED);
        jdbc.sql("UPDATE drivers SET assignment = 'NONE', current_ride_id = NULL WHERE id = :id")
                .param("id", driver)
                .update();

        Api.Response offline = api.put("/drivers/" + driver + "/availability", Map.of("status", "OFFLINE"));

        assertThat(offline.status()).as("the driver's copy has not seen the offer yet").isEqualTo(200);
        eventually().untilAsserted(() -> {
            assertThat(offerRows(rideId)).containsExactly(new OfferRow(driver, "EXPIRED"));
            assertThat(rideRow(rideId).status()).isEqualTo(RideStatus.MATCHING);
            assertThat(supplyRow(driver).online()).isFalse();
            assertThat(supplyRow(driver).assignment()).isEqualTo("NONE");
        });
    }

    @Test
    void cancellingAPendingOfferReleasesTheDriver() {
        Api.Response response = api.post("/rides/" + rideId + "/cancel");

        assertThat(response.status()).as(response.body()).isEqualTo(200);
        assertThat(response.as(RideResponse.class).status()).isEqualTo(RideStatus.CANCELLED);
        eventually().untilAsserted(() -> {
            assertThat(offerRows(rideId)).containsExactly(new OfferRow(driver, "CANCELLED"));
            assertThat(driverRow(driver).status()).isEqualTo(DriverStatus.AVAILABLE);
        });
    }

    @Test
    void cancellingAnAcceptedRideReleasesTheDriver() {
        acceptAndAwait(rideId, driver);

        Api.Response response = api.post("/rides/" + rideId + "/cancel");

        assertThat(response.status()).as(response.body()).isEqualTo(200);
        assertThat(rideRow(rideId).status()).isEqualTo(RideStatus.CANCELLED);
        awaitDriverStatus(driver, DriverStatus.AVAILABLE);
    }

    @Test
    void completingAnAcceptedRideReleasesTheDriver() {
        acceptAndAwait(rideId, driver);

        Api.Response response = api.post("/rides/" + rideId + "/complete");

        assertThat(response.status()).as(response.body()).isEqualTo(200);
        assertThat(rideRow(rideId).status()).isEqualTo(RideStatus.COMPLETED);
        assertThat(rideRow(rideId).assignedDriverId()).isEqualTo(driver);
        awaitDriverStatus(driver, DriverStatus.AVAILABLE);
    }

    @SuppressWarnings("unchecked")
    private String offerStatus(Api.Response response) {
        return String.valueOf(objectMapper.readValue(response.body(), Map.class).get("status"));
    }
}
