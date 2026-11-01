package com.ridehailing.acceptance;

import static org.assertj.core.api.Assertions.assertThat;

import com.ridehailing.driver.domain.DriverStatus;
import com.ridehailing.ride.domain.RideStatus;
import com.ridehailing.support.Api;
import com.ridehailing.support.IntegrationTestSupport;
import com.ridehailing.support.TestCity;
import java.util.Map;
import java.util.UUID;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;

@Tag("AC-02")
class AC02DriverAvailabilityAcceptanceIT extends IntegrationTestSupport {

    private final TestCity city = TestCity.next();

    @AfterEach
    void invariantsHold() {
        assertInvariants();
    }

    @Test
    void driverSwitchesBetweenOfflineAndAvailable() {
        String driver = fixtures.driver(DriverStatus.OFFLINE, city.centre());

        assertThat(setAvailability(driver, "AVAILABLE").status()).isEqualTo(200);
        assertThat(driverRow(driver).status()).isEqualTo(DriverStatus.AVAILABLE);
        eventually().untilAsserted(() -> assertThat(supplyRow(driver).online()).isTrue());

        assertThat(setAvailability(driver, "OFFLINE").status()).isEqualTo(200);
        assertThat(driverRow(driver).status()).isEqualTo(DriverStatus.OFFLINE);
        eventually().untilAsserted(() -> assertThat(supplyRow(driver).online()).isFalse());
    }

    @Test
    void repeatingTheSameAvailabilityIsHarmless() {
        String driver = fixtures.driver(DriverStatus.AVAILABLE, city.centre());

        assertThat(setAvailability(driver, "AVAILABLE").status()).isEqualTo(200);
        assertThat(driverRow(driver).status()).isEqualTo(DriverStatus.AVAILABLE);
    }

    @Test
    void onlyAvailableDriversReceiveOffers() {
        String offline = fixtures.driver(DriverStatus.OFFLINE, city.kmNorth(0.3));
        UUID rideId = createRide(fixtures.passenger(), city.centre()).id();

        awaitRideStatus(rideId, RideStatus.MATCHING);

        assertThat(offerRows(rideId)).isEmpty();
        assertThat(driverRow(offline).status()).isEqualTo(DriverStatus.OFFLINE);
    }

    @Test
    void anOfferedDriverBecomesOfferedAndCannotGoOffline() {
        String driver = fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(0.5));
        UUID rideId = createRide(fixtures.passenger(), city.centre()).id();
        awaitDriverStatus(driver, DriverStatus.OFFERED);

        Api.Response response = setAvailability(driver, "OFFLINE");

        assertThat(response.status()).as(response.body()).isEqualTo(409);
        assertThat(response.code()).isEqualTo("INVALID_DRIVER_TRANSITION");
        assertThat(driverRow(driver).status()).isEqualTo(DriverStatus.OFFERED);
        assertThat(driverRow(driver).currentRideId()).isEqualTo(rideId);
    }

    @Test
    void aBusyDriverCannotGoOfflineOrBeMadeAvailable() {
        String driver = fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(0.5));
        UUID rideId = createRide(fixtures.passenger(), city.centre()).id();
        awaitOfferedTo(rideId, driver);
        acceptAndAwait(rideId, driver);
        awaitDriverStatus(driver, DriverStatus.BUSY);

        assertThat(setAvailability(driver, "OFFLINE").status()).isEqualTo(409);
        assertThat(setAvailability(driver, "AVAILABLE").status()).isEqualTo(409);
        assertThat(driverRow(driver).status()).isEqualTo(DriverStatus.BUSY);
    }

    @Test
    void completingReturnsABusyDriverToAvailable() {
        String driver = fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(0.5));
        UUID rideId = createRide(fixtures.passenger(), city.centre()).id();
        awaitOfferedTo(rideId, driver);
        acceptAndAwait(rideId, driver);

        assertThat(api.post("/rides/" + rideId + "/complete").status()).isEqualTo(200);

        awaitDriverStatus(driver, DriverStatus.AVAILABLE);
        assertThat(supplyRow(driver).assignment()).isEqualTo("NONE");
    }

    @Test
    void availabilityApiCannotSetOfferedOrBusy() {
        String driver = fixtures.driver(DriverStatus.AVAILABLE, city.centre());

        Api.Response response = setAvailability(driver, "OFFERED");

        assertThat(response.status()).as(response.body()).isEqualTo(400);
        assertThat(response.code()).isEqualTo("UNSUPPORTED_AVAILABILITY");
    }

    @Test
    void rejectsUnknownStatusesAndDrivers() {
        String driver = fixtures.driver(DriverStatus.OFFLINE, city.centre());

        assertThat(setAvailability(driver, "SLEEPING").status()).isEqualTo(400);
        Api.Response unknown = setAvailability("D-does-not-exist", "AVAILABLE");
        assertThat(unknown.status()).isEqualTo(404);
        assertThat(unknown.code()).isEqualTo("DRIVER_NOT_FOUND");
    }

    private Api.Response setAvailability(String driverId, String status) {
        return api.put("/drivers/" + driverId + "/availability", Map.of("status", status));
    }
}
