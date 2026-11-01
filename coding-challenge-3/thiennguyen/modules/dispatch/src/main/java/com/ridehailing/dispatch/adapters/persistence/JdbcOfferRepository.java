package com.ridehailing.dispatch.adapters.persistence;

import com.ridehailing.common.jdbc.Timestamps;
import com.ridehailing.dispatch.application.port.OfferRepository;
import com.ridehailing.dispatch.domain.OfferStatus;
import com.ridehailing.dispatch.domain.OfferTransition;
import com.ridehailing.dispatch.domain.RideOffer;
import lombok.RequiredArgsConstructor;
import org.springframework.jdbc.core.simple.JdbcClient;
import org.springframework.stereotype.Repository;

import java.sql.ResultSet;
import java.sql.SQLException;
import java.time.Instant;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Repository
@RequiredArgsConstructor
public class JdbcOfferRepository implements OfferRepository {

    private static final String COLUMNS =
            "id, ride_id, driver_id, status, distance_km, created_at, expires_at, responded_at";

    private final JdbcClient jdbc;

    @Override
    public void insert(RideOffer offer) {
        jdbc.sql("""
                        INSERT INTO ride_offers (id, ride_id, driver_id, status, distance_km, created_at, expires_at)
                        VALUES (:id, :rideId, :driverId, :status, :distanceKm, :createdAt, :expiresAt)
                        """)
                .param("id", offer.id())
                .param("rideId", offer.rideId())
                .param("driverId", offer.driverId())
                .param("status", offer.status().name())
                .param("distanceKm", offer.distanceKm())
                .param("createdAt", Timestamps.toDb(offer.createdAt()))
                .param("expiresAt", Timestamps.toDb(offer.expiresAt()))
                .update();
    }

    @Override
    public boolean accept(UUID rideId, String driverId, Instant now) {
        return jdbc.sql("""
                        UPDATE ride_offers SET status = :to, responded_at = :now
                         WHERE ride_id = :rideId AND driver_id = :driverId AND status IN (:from) AND expires_at > :now
                        """)
                .param("to", OfferTransition.ACCEPT.to().name())
                .param("from", OfferTransition.ACCEPT.fromNames())
                .param("now", Timestamps.toDb(now))
                .param("rideId", rideId)
                .param("driverId", driverId)
                .update() == 1;
    }

    @Override
    public boolean resolvePending(UUID rideId, String driverId, OfferTransition transition, Instant now) {
        return jdbc.sql("""
                        UPDATE ride_offers SET status = :to, responded_at = :now
                         WHERE ride_id = :rideId AND driver_id = :driverId AND status IN (:from)
                        """)
                .param("to", transition.to().name())
                .param("from", transition.fromNames())
                .param("now", Timestamps.toDb(now))
                .param("rideId", rideId)
                .param("driverId", driverId)
                .update() == 1;
    }

    @Override
    public boolean expireIfDue(UUID offerId, Instant now) {
        return jdbc.sql("""
                        UPDATE ride_offers SET status = :to, responded_at = :now
                         WHERE id = :id AND status IN (:from) AND expires_at <= :now
                        """)
                .param("to", OfferTransition.EXPIRE.to().name())
                .param("from", OfferTransition.EXPIRE.fromNames())
                .param("now", Timestamps.toDb(now))
                .param("id", offerId)
                .update() == 1;
    }

    @Override
    public List<RideOffer> findDueForExpiry(Instant now, int limit) {
        return jdbc.sql("SELECT " + COLUMNS + """
                         FROM ride_offers
                        WHERE status = 'PENDING' AND expires_at <= :now
                        ORDER BY expires_at
                        LIMIT :limit
                        """)
                .param("now", Timestamps.toDb(now))
                .param("limit", limit)
                .query(JdbcOfferRepository::map)
                .list();
    }

    @Override
    public Optional<RideOffer> findPending(UUID rideId) {
        return jdbc.sql("SELECT " + COLUMNS + " FROM ride_offers WHERE ride_id = :rideId AND status = 'PENDING'")
                .param("rideId", rideId)
                .query(JdbcOfferRepository::map)
                .optional();
    }

    @Override
    public Optional<RideOffer> findLatest(UUID rideId, String driverId) {
        return jdbc.sql("SELECT " + COLUMNS + """
                         FROM ride_offers
                        WHERE ride_id = :rideId AND driver_id = :driverId
                        ORDER BY created_at DESC, id DESC
                        LIMIT 1
                        """)
                .param("rideId", rideId)
                .param("driverId", driverId)
                .query(JdbcOfferRepository::map)
                .optional();
    }

    @Override
    public List<RideOffer> findByRide(UUID rideId) {
        return jdbc.sql("SELECT " + COLUMNS + " FROM ride_offers WHERE ride_id = :rideId ORDER BY created_at, id")
                .param("rideId", rideId)
                .query(JdbcOfferRepository::map)
                .list();
    }

    private static RideOffer map(ResultSet resultSet, int rowNumber) throws SQLException {
        return new RideOffer(
                resultSet.getObject("id", UUID.class),
                resultSet.getObject("ride_id", UUID.class),
                resultSet.getString("driver_id"),
                OfferStatus.valueOf(resultSet.getString("status")),
                resultSet.getDouble("distance_km"),
                Timestamps.fromDb(resultSet, "created_at"),
                Timestamps.fromDb(resultSet, "expires_at"),
                Timestamps.fromDb(resultSet, "responded_at"));
    }
}
