package com.ridehailing.ride.application.port;

import java.time.Instant;
import java.util.Optional;
import java.util.UUID;

/**
 * Durable Idempotency-Key records, committed together with the ride they created.
 */
public interface IdempotencyRepository {

    /**
     * Claims the key for a new ride. A concurrent duplicate blocks on the key until the first transaction finishes,
     * then gets {@code false} and replays.
     */
    boolean claim(String key, String requestHash, UUID rideId, Instant now);

    Optional<IdempotencyRecord> find(String key);

    record IdempotencyRecord(String requestHash, UUID rideId) {
    }
}
