package com.ridehailing.integration;

import static org.assertj.core.api.Assertions.assertThat;

import com.ridehailing.contracts.Topics;
import com.ridehailing.driver.domain.DriverStatus;
import com.ridehailing.support.Api;
import com.ridehailing.support.IntegrationTestSupport;
import com.ridehailing.support.TestCity;
import io.micrometer.core.instrument.Counter;
import io.micrometer.core.instrument.MeterRegistry;
import io.micrometer.core.instrument.Timer;
import java.time.Duration;
import java.util.UUID;
import java.util.concurrent.TimeUnit;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.kafka.core.KafkaTemplate;

/**
 * The application meters follow what happens to rides and events. Meters are shared by every test in the context, so
 * each assertion compares a count before and after the test's own actions.
 */
class MetricsIT extends IntegrationTestSupport {

    @Autowired
    private MeterRegistry meters;

    @Autowired
    private KafkaTemplate<String, String> kafka;

    private final TestCity city = TestCity.next();

    @AfterEach
    void invariantsHold() {
        assertInvariants();
    }

    @Test
    void dispatchMetersFollowAFirstOfferARejectionARematchAndAnAcceptance() {
        String nearest = fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(0.5));
        String next = fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(1.0));
        double offered = count("ridehailing.dispatch.attempts", "outcome", "offered");
        long first = timerCount("ridehailing.dispatch.time.to.offer", "offer", "first");
        long rematch = timerCount("ridehailing.dispatch.time.to.offer", "offer", "rematch");
        double rejected = count("ridehailing.offers.ended", "result", "rejected");
        double accepted = count("ridehailing.offers.ended", "result", "accepted");
        long rideRequestedDeliveries = timerCount("ridehailing.events.latency", "consumer", "dispatch",
                "event", "RideRequested");

        UUID rideId = createRide(fixtures.passenger(), city.centre()).id();
        awaitOfferedTo(rideId, nearest);
        // At least: the sweepers may also offer rides other tests left waiting. Rejections and acceptances come only
        // from this test's own calls, so those are exact.
        eventually().untilAsserted(() -> {
            assertThat(timerCount("ridehailing.dispatch.time.to.offer", "offer", "first"))
                    .isGreaterThanOrEqualTo(first + 1);
            assertThat(timerCount("ridehailing.events.latency", "consumer", "dispatch", "event", "RideRequested"))
                    .isGreaterThan(rideRequestedDeliveries);
        });

        assertThat(reject(rideId, nearest).status()).isEqualTo(200);
        awaitOfferedTo(rideId, next);
        acceptAndAwait(rideId, next);

        eventually().untilAsserted(() -> {
            assertThat(count("ridehailing.dispatch.attempts", "outcome", "offered"))
                    .isGreaterThanOrEqualTo(offered + 2);
            assertThat(timerCount("ridehailing.dispatch.time.to.offer", "offer", "rematch"))
                    .isGreaterThanOrEqualTo(rematch + 1);
            assertThat(count("ridehailing.offers.ended", "result", "rejected")).isEqualTo(rejected + 1);
            assertThat(count("ridehailing.offers.ended", "result", "accepted")).isEqualTo(accepted + 1);
        });
    }

    @Test
    void aRideWithNoDriverNearbyCountsAsAnAttemptWithNoDriver() {
        double noDriver = count("ridehailing.dispatch.attempts", "outcome", "no_driver");

        createRide(fixtures.passenger(), city.centre());

        eventually().untilAsserted(() -> assertThat(count("ridehailing.dispatch.attempts", "outcome", "no_driver"))
                .isGreaterThan(noDriver));
    }

    @Test
    void timedOutAndCancelledOffersAreCountedByHowTheyEnded() {
        String driver = fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(0.5));
        double timedOut = count("ridehailing.offers.ended", "result", "timed_out");
        double cancelled = count("ridehailing.offers.ended", "result", "cancelled");
        UUID expiring = createRide(fixtures.passenger(), city.centre()).id();
        awaitOfferedTo(expiring, driver);

        clock.advance(Duration.ofSeconds(31));

        eventually().untilAsserted(() -> assertThat(count("ridehailing.offers.ended", "result", "timed_out"))
                .isGreaterThan(timedOut));
        UUID cancelling = createRide(fixtures.passenger(), city.centre()).id();
        awaitOfferedTo(cancelling, driver);
        Api.Response cancel = api.post("/rides/" + cancelling + "/cancel");
        assertThat(cancel.status()).as(cancel.body()).isEqualTo(200);
        eventually().untilAsserted(() -> assertThat(count("ridehailing.offers.ended", "result", "cancelled"))
                .isEqualTo(cancelled + 1));
    }

    @Test
    void aDeadLetteredRecordIsCountedByItsTopic() throws Exception {
        double deadLetters = deadLetters();

        kafka.send(Topics.RIDE_EVENTS, "poison-" + UUID.randomUUID(), "this is not an event")
                .get(10, TimeUnit.SECONDS);

        eventually().untilAsserted(() -> assertThat(deadLetters()).isEqualTo(deadLetters + 1));
    }

    @Test
    void aRideRequestedAgainBySweeperIsCounted() {
        String driver = fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(0.5));
        double republished = count("ridehailing.rides.republished");

        UUID rideId = fixtures.rideWithLostRideRequested(fixtures.passenger(), city.centre(), DESTINATION,
                clock.instant().minus(Duration.ofMinutes(10)));

        awaitOfferedTo(rideId, driver);
        assertThat(count("ridehailing.rides.republished")).isGreaterThan(republished);
    }

    private double deadLetters() {
        return count("ridehailing.events.dead.letters", "topic", Topics.RIDE_EVENTS);
    }

    private double count(String name, String... tags) {
        Counter counter = meters.find(name).tags(tags).counter();
        return counter == null ? 0 : counter.count();
    }

    private long timerCount(String name, String... tags) {
        Timer timer = meters.find(name).tags(tags).timer();
        return timer == null ? 0 : timer.count();
    }
}
