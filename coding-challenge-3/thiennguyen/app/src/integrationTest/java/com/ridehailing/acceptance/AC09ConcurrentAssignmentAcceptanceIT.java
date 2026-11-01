package com.ridehailing.acceptance;

import static org.assertj.core.api.Assertions.assertThat;

import com.ridehailing.dispatch.application.DispatchOutcome;
import com.ridehailing.dispatch.application.DispatchService;
import com.ridehailing.driver.domain.DriverStatus;
import com.ridehailing.support.Api;
import com.ridehailing.support.IntegrationTestSupport;
import com.ridehailing.support.TestCity;
import java.time.Duration;
import java.time.Instant;
import java.util.List;
import java.util.UUID;
import java.util.concurrent.Callable;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.RepeatedTest;
import org.junit.jupiter.api.Tag;
import org.springframework.beans.factory.annotation.Autowired;

/**
 * Two dispatch workers race for the same driver. The race is decided inside dispatch's own tables , so the test drives
 * dispatch directly: requests are inserted as WAITING without an event and dated in the future, so only the two workers
 * in the test touch them.
 */
@Tag("AC-09")
@Tag("concurrency")
class AC09ConcurrentAssignmentAcceptanceIT extends IntegrationTestSupport {

    @Autowired
    private DispatchService dispatch;

    @AfterEach
    void invariantsHold() {
        assertInvariants();
    }

    @RepeatedTest(20)
    void twoRidesCompetingForOneDriverProduceExactlyOneOffer() throws Exception {
        TestCity city = TestCity.next();
        String driver = fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(0.5));
        UUID first = waitingRequest(city);
        UUID second = waitingRequest(city);

        List<DispatchOutcome> outcomes = ConcurrentCalls.runTogether(List.of(
                () -> dispatch.dispatch(first), () -> dispatch.dispatch(second)));

        assertThat(outcomes).containsExactlyInAnyOrder(DispatchOutcome.OFFERED, DispatchOutcome.NO_DRIVER);
        UUID winner = outcomes.get(0) == DispatchOutcome.OFFERED ? first : second;
        UUID loser = winner.equals(first) ? second : first;
        assertThat(supplyRow(driver).assignment()).isEqualTo("OFFERED");
        assertThat(supplyRow(driver).currentRideId()).isEqualTo(winner);
        assertThat(requestRow(winner)).isEqualTo(new RequestRow("OFFERED", driver));
        assertThat(requestRow(loser)).isEqualTo(new RequestRow("WAITING", null));
        assertThat(pendingOffersFor(driver)).isEqualTo(1);
    }

    @RepeatedTest(10)
    void theLosingRideFallsThroughToTheNextDriver() throws Exception {
        TestCity city = TestCity.next();
        String nearest = fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(0.5));
        String next = fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(1.5));
        UUID first = waitingRequest(city);
        UUID second = waitingRequest(city);

        List<DispatchOutcome> outcomes = ConcurrentCalls.runTogether(List.of(
                () -> dispatch.dispatch(first), () -> dispatch.dispatch(second)));

        assertThat(outcomes).containsOnly(DispatchOutcome.OFFERED);
        assertThat(List.of(requestRow(first).driverId(), requestRow(second).driverId()))
                .containsExactlyInAnyOrder(nearest, next);
    }

    @RepeatedTest(10)
    void acceptAndRejectRacingForOneOfferHaveOneWinner() throws Exception {
        TestCity city = TestCity.next();
        String driver = fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(0.5));
        UUID rideId = waitingRequest(city);
        assertThat(dispatch.dispatch(rideId)).isEqualTo(DispatchOutcome.OFFERED);

        List<Callable<Api.Response>> calls = List.of(() -> accept(rideId, driver), () -> reject(rideId, driver));
        List<Api.Response> responses = ConcurrentCalls.runTogether(calls);

        assertThat(responses).extracting(Api.Response::status).containsExactlyInAnyOrder(200, 409);
        String outcome = requestRow(rideId).status();
        assertThat(outcome).isIn("ACCEPTED", "WAITING");
        assertThat(supplyRow(driver).assignment()).isEqualTo(outcome.equals("ACCEPTED") ? "BUSY" : "NONE");
    }

    private UUID waitingRequest(TestCity city) {
        Instant outOfTheSweepersReach = clock.instant().plus(Duration.ofMinutes(5));
        return fixtures.waitingDispatchRequest(city.centre(), outOfTheSweepersReach);
    }

    private long pendingOffersFor(String driverId) {
        return jdbc.sql("SELECT count(*) FROM ride_offers WHERE driver_id = :id AND status = 'PENDING'")
                .param("id", driverId)
                .query(Long.class)
                .single();
    }
}
