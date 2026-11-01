package com.ridehailing.ride.adapters.cache;

import com.ridehailing.common.cache.CacheProperties;
import com.ridehailing.common.cache.VersionedCache;
import com.ridehailing.ride.application.port.IdempotencyCache;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Component;

import java.util.Optional;
import java.util.UUID;

/**
 * Replay cache in front of {@code idempotency_keys}; entries exist only for committed keys and never change.
 */
@Component
@RequiredArgsConstructor
public class RedisIdempotencyCache implements IdempotencyCache {

    private static final String NAME = "idempotency";
    private static final long IMMUTABLE_VERSION = 1L;

    private final VersionedCache cache;
    private final CacheProperties properties;

    @Override
    public Optional<Entry> get(String idempotencyKey) {
        return cache.get(NAME, key(idempotencyKey)).flatMap(RedisIdempotencyCache::decode);
    }

    @Override
    public void putAfterCommit(String idempotencyKey, String requestHash, UUID rideId) {
        cache.putIfNewerAfterCommit(NAME, key(idempotencyKey), IMMUTABLE_VERSION, requestHash + ":" + rideId,
                properties.idempotencyTtl());
    }

    private static Optional<Entry> decode(String value) {
        int separator = value.indexOf(':');
        if (separator < 0) {
            return Optional.empty();
        }
        try {
            return Optional.of(new Entry(value.substring(0, separator), UUID.fromString(value.substring(separator + 1))));
        } catch (IllegalArgumentException exception) {
            return Optional.empty();
        }
    }

    private static String key(String idempotencyKey) {
        return "rh:idem:" + idempotencyKey;
    }
}
