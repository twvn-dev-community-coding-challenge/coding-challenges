package com.ridehailing.ride.application;

import com.ridehailing.common.events.OutboxWriter;
import com.ridehailing.common.logging.LogContext;
import com.ridehailing.contracts.EventType;
import com.ridehailing.contracts.GeoPoint;
import com.ridehailing.contracts.RideRequestedV1;
import com.ridehailing.ride.application.port.RideRepository;
import com.ridehailing.ride.domain.Ride;
import io.micrometer.core.instrument.MeterRegistry;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;
import org.springframework.transaction.support.TransactionTemplate;

import java.time.Clock;
import java.util.List;

/**
 * Self-healing for rides dispatch never picked up: a ride still REQUESTED well after its RideRequested was published
 * means the event was dead-lettered or lost before dispatch opened a request, and dispatch's own sweeper cannot see it.
 * Publishes RideRequested again through the outbox; dispatch opens a request at most once per ride, so a repeat that
 * crosses a slow MatchingStarted is harmless. {@code ridehailing.rides.republished} counts the rides published again.
 */
@Component
@RequiredArgsConstructor
@Slf4j
public class UndispatchedRideSweeper {

    private final RideRepository rides;
    private final OutboxWriter outbox;
    private final TransactionTemplate transactions;
    private final RideProperties properties;
    private final MeterRegistry meters;
    private final Clock clock;

    @Scheduled(fixedDelayString = "${ride.undispatched-sweep-interval-ms:5000}")
    public void sweep() {
        try {
            Integer republished = transactions.execute(status -> republishDue());
            if (republished != null && republished > 0) {
                meters.counter("ridehailing.rides.republished").increment(republished);
            }
        } catch (RuntimeException exception) {
            log.warn("Undispatched ride sweep failed; it will run again: {}", exception.toString());
        }
    }

    private int republishDue() {
        List<Ride> due = rides.claimUndispatched(
                clock.instant().minus(properties.undispatchedAfter()), properties.undispatchedBatchSize());
        for (Ride ride : due) {
            try (LogContext ignored = LogContext.ride(ride.id())) {
                GeoPoint pickup = new GeoPoint(ride.pickup().latitude(), ride.pickup().longitude());
                outbox.appendForRide(EventType.RIDE_REQUESTED, ride.id(), null,
                        new RideRequestedV1(ride.id(), ride.passengerId(), pickup));
                log.warn("Ride still REQUESTED {} after RideRequested; published it again",
                        properties.undispatchedAfter());
            }
        }
        return due.size();
    }
}
