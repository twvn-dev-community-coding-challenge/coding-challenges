package com.ridehailing.acceptance;

import static org.assertj.core.api.Assertions.assertThat;

import com.ridehailing.common.kernel.geo.Location;
import com.ridehailing.driver.domain.DriverStatus;
import com.ridehailing.support.Api;
import com.ridehailing.support.IntegrationTestSupport;
import com.ridehailing.support.TestCity;
import java.util.Map;
import java.util.UUID;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;

@Tag("AC-03")
class AC03DriverLocationAcceptanceIT extends IntegrationTestSupport {

    private final TestCity city = TestCity.next();

    @Test
    void rejectsInvalidCoordinates() {
        String driver = fixtures.driver(DriverStatus.AVAILABLE, city.centre());

        Api.Response outOfRange = updateLocation(driver, Map.of("latitude", 91.0, "longitude", 0.0));
        Api.Response missing = updateLocation(driver, Map.of("latitude", 10.0));

        assertThat(outOfRange.status()).as(outOfRange.body()).isEqualTo(400);
        assertThat(outOfRange.code()).isEqualTo("VALIDATION_FAILED");
        assertThat(missing.status()).as(missing.body()).isEqualTo(400);
    }

    @Test
    void storesTheLatestLocation() {
        String driver = fixtures.driver(DriverStatus.AVAILABLE, city.centre());
        Location latest = city.kmNorth(1.0);

        updateLocation(driver, coordinates(city.kmNorth(0.5)));
        Api.Response response = updateLocation(driver, coordinates(latest));

        assertThat(response.status()).as(response.body()).isEqualTo(200);
        Location stored = jdbc.sql("SELECT latitude, longitude FROM drivers WHERE id = :id")
                .param("id", driver)
                .query((rs, n) -> new Location(rs.getDouble("latitude"), rs.getDouble("longitude")))
                .single();
        assertThat(stored).isEqualTo(latest);
        assertThat(api.get("/drivers/" + driver).body()).contains(Double.toString(latest.latitude()));
    }

    @Test
    void matchingUsesTheLatestKnownLocation() {
        String movedCloser = fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(5.0));
        fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(1.5));
        Location latest = city.kmNorth(0.5);
        updateLocation(movedCloser, coordinates(latest));
        eventually().untilAsserted(() -> assertThat(supplyRow(movedCloser).latitude()).isEqualTo(latest.latitude()));

        UUID rideId = createRide(fixtures.passenger(), city.centre()).id();

        awaitOfferedTo(rideId, movedCloser);
    }

    private Api.Response updateLocation(String driverId, Map<String, Object> body) {
        return api.put("/drivers/" + driverId + "/location", body);
    }

    private static Map<String, Object> coordinates(Location location) {
        return Map.of("latitude", location.latitude(), "longitude", location.longitude());
    }
}
