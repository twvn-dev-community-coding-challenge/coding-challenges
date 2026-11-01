package com.ridehailing.ride.application.port;

import java.util.Optional;
import java.util.UUID;

/**
 * Replay cache in front of {@link IdempotencyRepository}; it answers repeats but never decides.
 */
public interface IdempotencyCache {

    Optional<Entry> get(String idempotencyKey);

    /**
     * Stores the committed key once the surrounding transaction commits; entries never change afterwards.
     */
    void putAfterCommit(String idempotencyKey, String requestHash, UUID rideId);

    record Entry(String requestHash, UUID rideId) {
    }
}
