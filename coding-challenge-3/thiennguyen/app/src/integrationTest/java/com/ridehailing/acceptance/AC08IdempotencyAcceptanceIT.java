package com.ridehailing.acceptance;

import static org.assertj.core.api.Assertions.assertThat;

import com.ridehailing.contracts.Topics;
import com.ridehailing.driver.domain.DriverStatus;
import com.ridehailing.ride.adapters.web.model.RideResponse;
import com.ridehailing.ride.domain.RideStatus;
import com.ridehailing.support.Api;
import com.ridehailing.support.IntegrationTestSupport;
import com.ridehailing.support.TestCity;
import java.time.Duration;
import java.util.List;
import java.util.UUID;
import java.util.concurrent.Callable;
import java.util.concurrent.TimeUnit;
import java.util.stream.IntStream;
import org.awaitility.Awaitility;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.kafka.core.KafkaTemplate;

@Tag("AC-08")
class AC08IdempotencyAcceptanceIT extends IntegrationTestSupport {

    @Autowired
    private KafkaTemplate<String, String> kafka;

    private final TestCity city = TestCity.next();

    @AfterEach
    void invariantsHold() {
        assertInvariants();
    }

    @Test
    void concurrentDuplicateCreatesProduceOneRide() throws Exception {
        String passenger = fixtures.passenger();
        String key = "key-" + UUID.randomUUID();
        List<Callable<Api.Response>> calls = IntStream.range(0, 20)
                .<Callable<Api.Response>>mapToObj(i -> () -> api.post("/rides", rideRequest(passenger, city.centre()), key))
                .toList();

        List<Api.Response> responses = ConcurrentCalls.runTogether(calls);

        assertThat(responses).allSatisfy(response -> assertThat(response.status()).as(response.body()).isEqualTo(201));
        assertThat(responses.stream().map(response -> response.as(RideResponse.class).id()).distinct()).hasSize(1);
        assertThat(jdbc.sql("SELECT count(*) FROM rides WHERE passenger_id = :id").param("id", passenger)
                .query(Long.class).single()).isEqualTo(1);
    }

    @Test
    void duplicateAcceptanceLeavesOneValidAcceptance() {
        String driver = fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(0.5));
        UUID rideId = createRide(fixtures.passenger(), city.centre()).id();
        awaitOfferedTo(rideId, driver);

        Api.Response first = accept(rideId, driver);
        Api.Response second = accept(rideId, driver);

        assertThat(first.status()).isEqualTo(200);
        assertThat(second.status()).as(second.body()).isEqualTo(200);
        assertThat(second.body()).contains("\"status\":\"ACCEPTED\"");
        assertThat(outboxCount(rideId, "RideAccepted")).isEqualTo(1);
        assertThat(supplyRow(driver).assignment()).isEqualTo("BUSY");
        awaitRideStatus(rideId, RideStatus.DRIVER_ACCEPTED);
        awaitDriverStatus(driver, DriverStatus.BUSY);
    }

    @Test
    void duplicateRejectionIsHarmless() {
        String driver = fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(0.5));
        UUID rideId = createRide(fixtures.passenger(), city.centre()).id();
        awaitOfferedTo(rideId, driver);

        assertThat(reject(rideId, driver).status()).isEqualTo(200);
        assertThat(reject(rideId, driver).status()).isEqualTo(200);
        assertThat(outboxCount(rideId, "OfferRejected")).isEqualTo(1);
    }

    @Test
    void aRedeliveredEventIsProcessedOnlyOnce() throws Exception {
        UUID rideId = createRide(fixtures.passenger(), city.centre()).id();
        awaitRideStatus(rideId, RideStatus.MATCHING);
        String requested = jdbc.sql("""
                        SELECT CAST(payload AS text) FROM outbox WHERE aggregate_id = :id AND event_type = 'RideRequested'
                        """)
                .param("id", rideId.toString())
                .query(String.class)
                .single();
        eventually().untilAsserted(() -> assertThat(processedCount(rideId)).isEqualTo(1));
        String lateDriver = fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(0.5));

        kafka.send(Topics.RIDE_EVENTS, rideId.toString(), requested).get(10, TimeUnit.SECONDS);

        Awaitility.await().during(Duration.ofSeconds(3)).atMost(Duration.ofSeconds(10)).untilAsserted(() -> {
            assertThat(offerRows(rideId)).isEmpty();
            assertThat(driverRow(lateDriver).status()).isEqualTo(DriverStatus.AVAILABLE);
        });
        assertThat(processedCount(rideId)).isEqualTo(1);
    }

    private long processedCount(UUID rideId) {
        return jdbc.sql("""
                        SELECT count(*) FROM processed_events p JOIN outbox o ON o.event_id = p.event_id
                         WHERE o.aggregate_id = :id AND o.event_type = 'RideRequested'
                        """)
                .param("id", rideId.toString())
                .query(Long.class)
                .single();
    }
}
