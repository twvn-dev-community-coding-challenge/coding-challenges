package com.ridehailing.dispatch.adapters.persistence;

import com.ridehailing.common.jdbc.Timestamps;
import com.ridehailing.common.kernel.geo.Location;
import com.ridehailing.dispatch.application.port.DispatchRequestRepository;
import com.ridehailing.dispatch.domain.DispatchRequest;
import com.ridehailing.dispatch.domain.DispatchRequestStatus;
import com.ridehailing.dispatch.domain.DispatchRequestTransition;
import lombok.RequiredArgsConstructor;
import org.springframework.jdbc.core.simple.JdbcClient;
import org.springframework.stereotype.Repository;

import java.sql.ResultSet;
import java.sql.SQLException;
import java.time.Clock;
import java.time.Instant;
import java.time.OffsetDateTime;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

/**
 * Every status change is a compare-and-set on the request row, which also locks it (always before supply rows).
 */
@Repository
@RequiredArgsConstructor
public class JdbcDispatchRequestRepository implements DispatchRequestRepository {

    private static final String COLUMNS =
            "ride_id, pickup_latitude, pickup_longitude, status, driver_id, version, created_at, updated_at";

    private final JdbcClient jdbc;
    private final Clock clock;

    @Override
    public boolean open(UUID rideId, Location pickup) {
        return jdbc.sql("""
                        INSERT INTO dispatch_requests (ride_id, pickup_latitude, pickup_longitude, status, created_at,
                                                       updated_at)
                        VALUES (:rideId, :latitude, :longitude, 'WAITING', :now, :now)
                        ON CONFLICT (ride_id) DO NOTHING
                        """)
                .param("rideId", rideId)
                .param("latitude", pickup.latitude())
                .param("longitude", pickup.longitude())
                .param("now", now())
                .update() == 1;
    }

    @Override
    public boolean recordEnded(UUID rideId, DispatchRequestStatus outcome) {
        if (!outcome.isTerminal()) {
            throw new IllegalArgumentException("Not a terminal status: " + outcome);
        }
        return jdbc.sql("""
                        INSERT INTO dispatch_requests (ride_id, status, created_at, updated_at)
                        VALUES (:rideId, :outcome, :now, :now)
                        ON CONFLICT (ride_id) DO NOTHING
                        """)
                .param("rideId", rideId)
                .param("outcome", outcome.name())
                .param("now", now())
                .update() == 1;
    }

    @Override
    public Optional<DispatchRequest> find(UUID rideId) {
        return jdbc.sql("SELECT " + COLUMNS + " FROM dispatch_requests WHERE ride_id = :rideId")
                .param("rideId", rideId)
                .query(JdbcDispatchRequestRepository::map)
                .optional();
    }

    @Override
    public Optional<DispatchRequest> lock(UUID rideId) {
        return jdbc.sql("SELECT " + COLUMNS + " FROM dispatch_requests WHERE ride_id = :rideId FOR UPDATE")
                .param("rideId", rideId)
                .query(JdbcDispatchRequestRepository::map)
                .optional();
    }

    @Override
    public Optional<DispatchRequest> lockForMatching(UUID rideId) {
        return jdbc.sql("""
                        UPDATE dispatch_requests SET version = version + 1, updated_at = :now
                         WHERE ride_id = :rideId AND status IN (:from)
                        RETURNING\s""" + COLUMNS)
                .param("now", now())
                .param("rideId", rideId)
                .param("from", DispatchRequestTransition.OFFER.fromNames())
                .query(JdbcDispatchRequestRepository::map)
                .optional();
    }

    @Override
    public boolean apply(UUID rideId, DispatchRequestTransition transition, String driverId) {
        String heldDriver = switch (transition.heldDriver()) {
            case KEEP -> "";
            case SET -> ", driver_id = :driverId";
            case CLEAR -> ", driver_id = NULL";
        };
        String holder = transition.requiresHolder() ? " AND driver_id = :driverId" : "";
        JdbcClient.StatementSpec statement = jdbc.sql("""
                        UPDATE dispatch_requests SET status = :to, version = version + 1, updated_at = :now"""
                        + heldDriver + " WHERE ride_id = :rideId AND status IN (:from)" + holder)
                .param("to", transition.to().name())
                .param("now", now())
                .param("rideId", rideId)
                .param("from", transition.fromNames());
        if (transition.heldDriver() == DispatchRequestTransition.HeldDriver.SET || transition.requiresHolder()) {
            statement = statement.param("driverId", requireDriver(transition, driverId));
        }
        return statement.update() == 1;
    }

    @Override
    public List<UUID> findStaleWaiting(Instant updatedBefore, int limit) {
        return jdbc.sql("""
                        SELECT ride_id FROM dispatch_requests
                         WHERE status = 'WAITING' AND updated_at < :updatedBefore
                         ORDER BY updated_at
                         LIMIT :limit
                        """)
                .param("updatedBefore", Timestamps.toDb(updatedBefore))
                .param("limit", limit)
                .query((resultSet, rowNumber) -> resultSet.getObject("ride_id", UUID.class))
                .list();
    }

    private static String requireDriver(DispatchRequestTransition transition, String driverId) {
        if (driverId == null) {
            throw new IllegalArgumentException(transition + " needs a driver id");
        }
        return driverId;
    }

    private OffsetDateTime now() {
        return Timestamps.toDb(clock.instant());
    }

    private static DispatchRequest map(ResultSet resultSet, int rowNumber) throws SQLException {
        Double latitude = resultSet.getObject("pickup_latitude", Double.class);
        Double longitude = resultSet.getObject("pickup_longitude", Double.class);
        return new DispatchRequest(
                resultSet.getObject("ride_id", UUID.class),
                latitude == null || longitude == null ? null : new Location(latitude, longitude),
                DispatchRequestStatus.valueOf(resultSet.getString("status")),
                resultSet.getString("driver_id"),
                resultSet.getLong("version"),
                Timestamps.fromDb(resultSet, "created_at"),
                Timestamps.fromDb(resultSet, "updated_at"));
    }
}
