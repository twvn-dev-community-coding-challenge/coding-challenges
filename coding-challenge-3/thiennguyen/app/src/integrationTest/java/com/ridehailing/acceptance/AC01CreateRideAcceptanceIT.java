package com.ridehailing.acceptance;

import static org.assertj.core.api.Assertions.assertThat;

import com.ridehailing.ride.adapters.web.model.RideResponse;
import com.ridehailing.ride.domain.RideStatus;
import com.ridehailing.support.Api;
import com.ridehailing.support.IntegrationTestSupport;
import com.ridehailing.support.TestCity;
import java.util.Map;
import java.util.UUID;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.CsvSource;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.StringRedisTemplate;

@Tag("AC-01")
class AC01CreateRideAcceptanceIT extends IntegrationTestSupport {

    @Autowired
    private StringRedisTemplate redis;

    private final TestCity city = TestCity.next();
    private String passenger;

    @BeforeEach
    void setUp() {
        passenger = fixtures.passenger();
    }

    @Test
    void createsARideInRequestedThatCanBeRetrieved() {
        Api.Response response = api.post("/rides", rideRequest(passenger, city.centre()), "key-" + UUID.randomUUID());

        assertThat(response.status()).as(response.body()).isEqualTo(201);
        RideResponse ride = response.as(RideResponse.class);
        assertThat(ride.id()).isNotNull();
        assertThat(ride.status()).isEqualTo(RideStatus.REQUESTED);
        assertThat(ride.assignedDriverId()).isNull();
        assertThat(response.location()).isEqualTo("/rides/" + ride.id());
        assertThat(response.replayed()).isEqualTo("false");

        Api.Response fetched = api.get("/rides/" + ride.id());
        assertThat(fetched.status()).as(fetched.body()).isEqualTo(200);
        RideResponse retrieved = fetched.as(RideResponse.class);
        assertThat(retrieved.id()).isEqualTo(ride.id());
        assertThat(retrieved.passengerId()).isEqualTo(passenger);
        assertThat(retrieved.pickup().latitude()).isEqualTo(city.centre().latitude());
    }

    @Test
    void generatesAUniqueRideIdForEachRequest() {
        RideResponse first = createRide(passenger, city.centre());
        RideResponse second = createRide(passenger, city.centre());

        assertThat(first.id()).isNotEqualTo(second.id());
    }

    @ParameterizedTest(name = "latitude {0}, longitude {1}")
    @CsvSource({"90.0001, 0", "-90.0001, 0", "0, 180.0001", "0, -180.0001"})
    void rejectsInvalidCoordinates(double latitude, double longitude) {
        Map<String, Object> body = Map.of(
                "passengerId", passenger,
                "pickup", Map.of("latitude", latitude, "longitude", longitude),
                "destination", Map.of("latitude", 10.8, "longitude", 106.6));

        Api.Response response = api.post("/rides", body, "key-" + UUID.randomUUID());

        assertThat(response.status()).as(response.body()).isEqualTo(400);
        assertThat(response.code()).isEqualTo("VALIDATION_FAILED");
        assertThat(ridesOf(passenger)).isZero();
    }

    @Test
    void rejectsAMissingPassengerId() {
        Map<String, Object> body = Map.of(
                "pickup", Map.of("latitude", 1.0, "longitude", 1.0),
                "destination", Map.of("latitude", 1.1, "longitude", 1.1));

        Api.Response response = api.post("/rides", body, "key-" + UUID.randomUUID());

        assertThat(response.status()).as(response.body()).isEqualTo(400);
        assertThat(response.code()).isEqualTo("VALIDATION_FAILED");
    }

    @Test
    void rejectsAPassengerThatDoesNotExist() {
        Api.Response response = api.post("/rides", rideRequest("P-unknown", city.centre()), "key-" + UUID.randomUUID());

        assertThat(response.status()).as(response.body()).isEqualTo(400);
        assertThat(response.code()).isEqualTo("UNKNOWN_PASSENGER");
    }

    @Test
    void requiresAnIdempotencyKey() {
        Api.Response response = api.post("/rides", rideRequest(passenger, city.centre()), null);

        assertThat(response.status()).as(response.body()).isEqualTo(400);
        assertThat(response.code()).isEqualTo("IDEMPOTENCY_KEY_REQUIRED");
    }

    @Test
    void repeatingARequestWithTheSameKeyReturnsTheOriginalRide() {
        String key = "key-" + UUID.randomUUID();
        Api.Response first = api.post("/rides", rideRequest(passenger, city.centre()), key);
        Api.Response second = api.post("/rides", rideRequest(passenger, city.centre()), key);

        assertThat(second.status()).as(second.body()).isEqualTo(201);
        assertThat(second.replayed()).isEqualTo("true");
        UUID rideId = first.as(RideResponse.class).id();
        assertThat(second.as(RideResponse.class).id()).isEqualTo(rideId);
        assertThat(ridesOf(passenger)).isEqualTo(1);
        assertThat(outboxCount(rideId, "RideRequested")).isEqualTo(1);
    }

    @Test
    void reusingAKeyWithADifferentRequestIsRejected() {
        String key = "key-" + UUID.randomUUID();
        api.post("/rides", rideRequest(passenger, city.centre()), key);

        Api.Response response = api.post("/rides", rideRequest(passenger, city.kmNorth(1.0)), key);

        assertThat(response.status()).as(response.body()).isEqualTo(422);
        assertThat(response.code()).isEqualTo("IDEMPOTENCY_KEY_REUSED");
        assertThat(ridesOf(passenger)).isEqualTo(1);
    }

    @Test
    void replayStaysCorrectWhenTheCacheEntryIsLost() {
        String key = "key-" + UUID.randomUUID();
        UUID rideId = api.post("/rides", rideRequest(passenger, city.centre()), key).as(RideResponse.class).id();
        redis.delete("rh:idem:" + key);

        Api.Response replay = api.post("/rides", rideRequest(passenger, city.centre()), key);

        assertThat(replay.status()).as(replay.body()).isEqualTo(201);
        assertThat(replay.as(RideResponse.class).id()).isEqualTo(rideId);
        assertThat(ridesOf(passenger)).isEqualTo(1);
    }

    private long ridesOf(String passengerId) {
        return jdbc.sql("SELECT count(*) FROM rides WHERE passenger_id = :id")
                .param("id", passengerId)
                .query(Long.class)
                .single();
    }
}
