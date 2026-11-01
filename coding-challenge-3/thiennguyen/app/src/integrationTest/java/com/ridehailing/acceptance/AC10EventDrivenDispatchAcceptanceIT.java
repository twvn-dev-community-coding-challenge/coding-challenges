package com.ridehailing.acceptance;

import static org.assertj.core.api.Assertions.assertThat;

import com.ridehailing.common.events.EventCodec;
import com.ridehailing.contracts.EventEnvelope;
import com.ridehailing.contracts.EventType;
import com.ridehailing.contracts.GeoPoint;
import com.ridehailing.contracts.RideCancelledV1;
import com.ridehailing.contracts.RideRequestedV1;
import com.ridehailing.contracts.Topics;
import com.ridehailing.driver.domain.DriverStatus;
import com.ridehailing.support.IntegrationTestSupport;
import com.ridehailing.support.TestCity;
import java.time.Duration;
import java.util.List;
import java.util.Optional;
import java.util.UUID;
import java.util.concurrent.TimeUnit;
import java.util.function.Predicate;
import org.apache.kafka.clients.consumer.Consumer;
import org.apache.kafka.clients.consumer.ConsumerRecord;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.kafka.core.ConsumerFactory;
import org.springframework.kafka.core.KafkaTemplate;

@Tag("AC-10")
class AC10EventDrivenDispatchAcceptanceIT extends IntegrationTestSupport {

    @Autowired
    private KafkaTemplate<String, String> kafka;

    @Autowired
    private ConsumerFactory<String, String> consumers;

    @Autowired
    private EventCodec codec;

    private final TestCity city = TestCity.next();

    @Test
    void aNewRideIsDispatchedThroughTheBroker() {
        String driver = fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(0.5));

        UUID rideId = createRide(fixtures.passenger(), city.centre()).id();

        awaitOfferedTo(rideId, driver);
        assertThat(unpublishedEvents(rideId)).isZero();
        Optional<ConsumerRecord<String, String>> requested = findRecord(Topics.RIDE_EVENTS,
                record -> rideId.toString().equals(record.key()) && record.value().contains("\"RideRequested\""));
        assertThat(requested).as("RideRequested on %s keyed by rideId", Topics.RIDE_EVENTS).isPresent();
        assertThat(requested.get().value()).as("pickup travels in the event").contains("\"pickup\"");
        Optional<ConsumerRecord<String, String>> offered = findRecord(Topics.DISPATCH_EVENTS,
                record -> rideId.toString().equals(record.key()) && record.value().contains("\"RideOffered\""));
        assertThat(offered).as("RideOffered on %s keyed by rideId", Topics.DISPATCH_EVENTS).isPresent();
    }

    @Test
    void anUnprocessableEventIsDeadLetteredWithoutBlockingDispatch() throws Exception {
        String key = "poison-" + UUID.randomUUID();

        kafka.send(Topics.RIDE_EVENTS, key, "this is not an event").get(10, TimeUnit.SECONDS);

        String deadLetters = Topics.deadLetter(Topics.RIDE_EVENTS);
        assertThat(findRecord(deadLetters, record -> key.equals(record.key())))
                .as("poison record on %s", deadLetters).isPresent();
        String driver = fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(0.5));
        UUID rideId = createRide(fixtures.passenger(), city.centre()).id();
        awaitOfferedTo(rideId, driver);
    }

    @Test
    void theSweeperRecoversARideWhoseDispatchMessageWasLost() {
        String driver = fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(0.5));

        UUID rideId = fixtures.waitingDispatchRequest(city.centre(), clock.instant().minus(Duration.ofMinutes(1)));

        eventually().untilAsserted(() -> assertThat(requestRow(rideId)).isEqualTo(new RequestRow("OFFERED", driver)));
        assertThat(offerRows(rideId)).containsExactly(new OfferRow(driver, "PENDING"));
    }

    @Test
    void aRideWhoseRideRequestedNeverReachedDispatchIsRequestedAgain() {
        String driver = fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(0.5));

        UUID rideId = fixtures.rideWithLostRideRequested(fixtures.passenger(), city.centre(), DESTINATION,
                clock.instant().minus(Duration.ofMinutes(10)));

        awaitOfferedTo(rideId, driver);
        assertThat(requestRow(rideId)).isEqualTo(new RequestRow("OFFERED", driver));
        assertThat(outboxCount(rideId, "RideRequested")).as("published again exactly once").isEqualTo(1);
        assertInvariants();
    }

    @Test
    void aRideThatEndedBeforeDispatchOpenedItIsNeverOffered() throws Exception {
        String driver = fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(0.5));
        UUID rideId = UUID.randomUUID();

        UUID cancelled = publishRideEvent(EventType.RIDE_CANCELLED, rideId,
                new RideCancelledV1(rideId, "REQUESTED", null));
        awaitProcessedByDispatch(cancelled);
        assertThat(requestRow(rideId)).isEqualTo(new RequestRow("CANCELLED", null));

        UUID lateRequest = publishRideEvent(EventType.RIDE_REQUESTED, rideId, new RideRequestedV1(
                rideId, fixtures.passenger(), new GeoPoint(city.centre().latitude(), city.centre().longitude())));
        awaitProcessedByDispatch(lateRequest);

        assertThat(requestRow(rideId)).isEqualTo(new RequestRow("CANCELLED", null));
        assertThat(offerRows(rideId)).isEmpty();
        assertThat(supplyRow(driver).assignment()).isEqualTo("NONE");
        invariants.assertHold();
    }

    /** Publishes an event as ride would, bypassing ride's state (dispatch never reads it). */
    private UUID publishRideEvent(EventType type, UUID rideId, Object payload) throws Exception {
        EventEnvelope envelope = EventEnvelope.create(type, rideId.toString(), null, clock.instant(),
                codec.toPayload(payload));
        kafka.send(Topics.RIDE_EVENTS, rideId.toString(), codec.encode(envelope)).get(10, TimeUnit.SECONDS);
        return envelope.eventId();
    }

    private void awaitProcessedByDispatch(UUID eventId) {
        eventually().untilAsserted(() -> assertThat(jdbc.sql("""
                        SELECT count(*) FROM processed_events WHERE consumer = 'dispatch' AND event_id = :id
                        """)
                .param("id", eventId)
                .query(Long.class)
                .single()).isEqualTo(1));
    }

    private long unpublishedEvents(UUID rideId) {
        return jdbc.sql("SELECT count(*) FROM outbox WHERE aggregate_id = :id AND published_at IS NULL")
                .param("id", rideId.toString())
                .query(Long.class)
                .single();
    }

    private Optional<ConsumerRecord<String, String>> findRecord(
            String topic, Predicate<ConsumerRecord<String, String>> match) {
        try (Consumer<String, String> consumer = consumers.createConsumer("probe-" + UUID.randomUUID(), "probe")) {
            consumer.subscribe(List.of(topic));
            long deadline = System.nanoTime() + Duration.ofSeconds(30).toNanos();
            while (System.nanoTime() < deadline) {
                for (ConsumerRecord<String, String> record : consumer.poll(Duration.ofMillis(500))) {
                    if (match.test(record)) {
                        return Optional.of(record);
                    }
                }
            }
        }
        return Optional.empty();
    }
}
