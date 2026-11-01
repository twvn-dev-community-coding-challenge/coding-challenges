package com.ridehailing.ride.adapters.persistence;

import com.ridehailing.common.jdbc.Timestamps;
import com.ridehailing.ride.application.port.IdempotencyRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.jdbc.core.simple.JdbcClient;
import org.springframework.stereotype.Repository;

import java.time.Instant;
import java.util.Optional;
import java.util.UUID;

@Repository
@RequiredArgsConstructor
public class JdbcIdempotencyRepository implements IdempotencyRepository {

    private final JdbcClient jdbc;

    @Override
    public boolean claim(String key, String requestHash, UUID rideId, Instant now) {
        return jdbc.sql("""
                        INSERT INTO idempotency_keys (key, request_hash, ride_id, created_at)
                        VALUES (:key, :requestHash, :rideId, :createdAt)
                        ON CONFLICT (key) DO NOTHING
                        """)
                .param("key", key)
                .param("requestHash", requestHash)
                .param("rideId", rideId)
                .param("createdAt", Timestamps.toDb(now))
                .update() == 1;
    }

    @Override
    public Optional<IdempotencyRecord> find(String key) {
        return jdbc.sql("SELECT request_hash, ride_id FROM idempotency_keys WHERE key = :key")
                .param("key", key)
                .query((resultSet, rowNumber) -> new IdempotencyRecord(
                        resultSet.getString("request_hash"), resultSet.getObject("ride_id", UUID.class)))
                .optional();
    }
}
