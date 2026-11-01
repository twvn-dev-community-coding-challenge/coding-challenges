package com.ridehailing.dispatch.adapters.persistence;

import com.ridehailing.common.jdbc.Timestamps;
import com.ridehailing.dispatch.application.port.ExclusionRepository;
import com.ridehailing.dispatch.domain.OfferStatus;
import lombok.RequiredArgsConstructor;
import org.springframework.jdbc.core.simple.JdbcClient;
import org.springframework.stereotype.Repository;

import java.time.Instant;
import java.util.HashSet;
import java.util.Set;
import java.util.UUID;

@Repository
@RequiredArgsConstructor
public class JdbcExclusionRepository implements ExclusionRepository {

    private final JdbcClient jdbc;

    @Override
    public void add(UUID rideId, String driverId, OfferStatus reason, Instant now) {
        jdbc.sql("""
                        INSERT INTO ride_driver_exclusions (ride_id, driver_id, reason, created_at)
                        VALUES (:rideId, :driverId, :reason, :createdAt)
                        ON CONFLICT DO NOTHING
                        """)
                .param("rideId", rideId)
                .param("driverId", driverId)
                .param("reason", reason.name())
                .param("createdAt", Timestamps.toDb(now))
                .update();
    }

    @Override
    public Set<String> forRide(UUID rideId) {
        return new HashSet<>(jdbc.sql("SELECT driver_id FROM ride_driver_exclusions WHERE ride_id = :rideId")
                .param("rideId", rideId)
                .query(String.class)
                .list());
    }

    @Override
    public boolean contains(UUID rideId, String driverId) {
        return jdbc.sql("""
                        SELECT EXISTS (SELECT 1 FROM ride_driver_exclusions WHERE ride_id = :rideId AND driver_id = :driverId)
                        """)
                .param("rideId", rideId)
                .param("driverId", driverId)
                .query(Boolean.class)
                .single();
    }
}
