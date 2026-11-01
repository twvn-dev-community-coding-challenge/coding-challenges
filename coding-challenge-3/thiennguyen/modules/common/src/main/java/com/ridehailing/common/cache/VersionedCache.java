package com.ridehailing.common.cache;

import io.micrometer.core.instrument.MeterRegistry;
import java.time.Duration;
import java.util.List;
import java.util.Optional;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.data.redis.core.script.DefaultRedisScript;
import org.springframework.data.redis.core.script.RedisScript;
import org.springframework.stereotype.Component;
import org.springframework.transaction.support.TransactionSynchronization;
import org.springframework.transaction.support.TransactionSynchronizationManager;

/**
 * Redis read cache: values carry the row version and are only stored when newer than what is cached, writes happen
 * after the database commit, and every Redis failure falls back silently to the caller's source of truth.
 */
@Component
@RequiredArgsConstructor
@Slf4j
public class VersionedCache {

    private static final RedisScript<Long> PUT_IF_NEWER = new DefaultRedisScript<>("""
            local current = redis.call('HGET', KEYS[1], 'v')
            if current and tonumber(current) >= tonumber(ARGV[1]) then
              return 0
            end
            redis.call('HSET', KEYS[1], 'v', ARGV[1], 'd', ARGV[2])
            redis.call('PEXPIRE', KEYS[1], ARGV[3])
            return 1
            """, Long.class);

    private final StringRedisTemplate redis;
    private final MeterRegistry meters;

    public Optional<String> get(String cacheName, String key) {
        try {
            String value = redis.<String, String>opsForHash().get(key, "d");
            record(cacheName, value == null ? "miss" : "hit");
            return Optional.ofNullable(value);
        } catch (RuntimeException exception) {
            onError(cacheName, "get", exception);
            return Optional.empty();
        }
    }

    public boolean putIfNewer(String cacheName, String key, long version, String value, Duration ttl) {
        try {
            Long stored = redis.execute(
                    PUT_IF_NEWER, List.of(key), Long.toString(version), value, Long.toString(ttl.toMillis()));
            return stored != null && stored == 1L;
        } catch (RuntimeException exception) {
            onError(cacheName, "put", exception);
            return false;
        }
    }

    /** Stores the value once the surrounding transaction commits; a rolled-back transaction never reaches Redis. */
    public void putIfNewerAfterCommit(String cacheName, String key, long version, String value, Duration ttl) {
        if (!TransactionSynchronizationManager.isSynchronizationActive()) {
            putIfNewer(cacheName, key, version, value, ttl);
            return;
        }
        TransactionSynchronizationManager.registerSynchronization(new TransactionSynchronization() {
            @Override
            public void afterCommit() {
                putIfNewer(cacheName, key, version, value, ttl);
            }
        });
    }

    private void record(String cacheName, String result) {
        meters.counter("ridehailing.cache.requests", "cache", cacheName, "result", result).increment();
    }

    private void onError(String cacheName, String operation, RuntimeException exception) {
        record(cacheName, "error");
        log.warn("Cache {} {} failed, falling back to the database: {}", cacheName, operation, exception.toString());
    }
}
