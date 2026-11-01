package com.ridehailing.ride.application;

import com.ridehailing.common.error.UnprocessableException;
import com.ridehailing.common.error.ValidationException;
import com.ridehailing.common.events.OutboxWriter;
import com.ridehailing.common.logging.LogContext;
import com.ridehailing.contracts.EventType;
import com.ridehailing.contracts.GeoPoint;
import com.ridehailing.contracts.RideRequestedV1;
import com.ridehailing.ride.application.port.*;
import com.ridehailing.ride.domain.Ride;
import com.ridehailing.ride.domain.RideStatus;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.support.TransactionTemplate;

import java.time.Clock;
import java.time.Instant;
import java.util.UUID;

/**
 * Creates a ride exactly once per Idempotency-Key. The ride, its key and its RideRequested outbox event commit
 * together; the Redis replay cache is consulted first but never decides.
 */
@Service
@RequiredArgsConstructor
@Slf4j
public class RideCreationService {

    private final RideRepository rides;
    private final IdempotencyRepository keys;
    private final Passengers passengers;
    private final OutboxWriter outbox;
    private final RideReadCache rideCache;
    private final IdempotencyCache keyCache;
    private final TransactionTemplate transactions;
    private final Clock clock;

    public CreationResult create(String idempotencyKey, CreateRideCommand command) {
        String requestHash = RequestFingerprint.of(command);
        var cached = keyCache.get(idempotencyKey);
        if (cached.isPresent()) {
            return replay(cached.get().requestHash(), cached.get().rideId(), requestHash, "cache");
        }
        return transactions.execute(_ -> createOrReplay(idempotencyKey, requestHash, command));
    }

    private CreationResult createOrReplay(String idempotencyKey, String requestHash, CreateRideCommand command) {
        if (!passengers.exists(command.passengerId())) {
            throw new ValidationException(
                    "UNKNOWN_PASSENGER", "Passenger %s does not exist".formatted(command.passengerId()));
        }
        Instant now = clock.instant();
        UUID rideId = UUID.randomUUID();
        if (!keys.claim(idempotencyKey, requestHash, rideId, now)) {
            var existing = keys.find(idempotencyKey)
                    .orElseThrow(() -> new IllegalStateException("Claimed idempotency key vanished"));
            return replay(existing.requestHash(), existing.rideId(), requestHash, "database");
        }
        Ride ride = Ride.requested(rideId, command.passengerId(), command.pickup(), command.destination(), now);
        rides.insert(ride);
        try (LogContext ignored = LogContext.ride(rideId)) {
            rides.recordHistory(rideId, null, RideStatus.REQUESTED, null, "ride requested");
            GeoPoint pickup = new GeoPoint(command.pickup().latitude(), command.pickup().longitude());
            outbox.appendForRide(EventType.RIDE_REQUESTED, rideId, null,
                    new RideRequestedV1(rideId, command.passengerId(), pickup));
            rideCache.putAfterCommit(ride);
            keyCache.putAfterCommit(idempotencyKey, requestHash, rideId);
            log.info("Ride created in REQUESTED");
        }
        return new CreationResult(ride, false);
    }

    private CreationResult replay(String storedHash, UUID rideId, String requestHash, String source) {
        if (!storedHash.equals(requestHash)) {
            throw new UnprocessableException(
                    "IDEMPOTENCY_KEY_REUSED", "Idempotency-Key was already used with a different request");
        }
        Ride ride = rideCache.get(rideId)
                .or(() -> rides.find(rideId))
                .orElseThrow(() -> new IllegalStateException("Ride of a committed idempotency key is missing"));
        try (LogContext ignored = LogContext.ride(rideId)) {
            log.info("Duplicate create request answered from {} without creating a ride", source);
        }
        return new CreationResult(ride, true);
    }

    public record CreationResult(Ride ride, boolean replayed) {
    }
}
