package com.ridehailing.ride.adapters.persistence;

import com.ridehailing.common.jdbc.Timestamps;
import com.ridehailing.common.kernel.geo.Location;
import com.ridehailing.ride.application.port.RideRepository;
import com.ridehailing.ride.domain.Ride;
import com.ridehailing.ride.domain.RideHistoryEntry;
import com.ridehailing.ride.domain.RideStatus;
import com.ridehailing.ride.domain.RideTransition;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.sql.Types;
import java.time.Clock;
import java.time.Instant;
import java.util.List;
import java.util.Optional;
import java.util.UUID;
import lombok.RequiredArgsConstructor;
import org.springframework.jdbc.core.simple.JdbcClient;
import org.springframework.stereotype.Repository;

@Repository
@RequiredArgsConstructor
public class JdbcRideRepository implements RideRepository {

    private static final String COLUMNS = """
            id, passenger_id, pickup_latitude, pickup_longitude, destination_latitude, destination_longitude,
            status, assigned_driver_id, version, created_at, updated_at""";
    private static final String QUALIFIED_COLUMNS = """
            r.id, r.passenger_id, r.pickup_latitude, r.pickup_longitude, r.destination_latitude,
            r.destination_longitude, r.status, r.assigned_driver_id, r.version, r.created_at, r.updated_at""";

    private final JdbcClient jdbc;
    private final Clock clock;

    @Override
    public void insert(Ride ride) {
        jdbc.sql("""
                        INSERT INTO rides (id, passenger_id, pickup_latitude, pickup_longitude, destination_latitude,
                                           destination_longitude, status, version, created_at, updated_at,
                                           dispatch_requested_at)
                        VALUES (:id, :passengerId, :pickupLatitude, :pickupLongitude, :destinationLatitude,
                                :destinationLongitude, :status, :version, :createdAt, :updatedAt, :createdAt)
                        """)
                .param("id", ride.id())
                .param("passengerId", ride.passengerId())
                .param("pickupLatitude", ride.pickup().latitude())
                .param("pickupLongitude", ride.pickup().longitude())
                .param("destinationLatitude", ride.destination().latitude())
                .param("destinationLongitude", ride.destination().longitude())
                .param("status", ride.status().name())
                .param("version", ride.version())
                .param("createdAt", Timestamps.toDb(ride.createdAt()))
                .param("updatedAt", Timestamps.toDb(ride.updatedAt()))
                .update();
    }

    @Override
    public Optional<Ride> find(UUID rideId) {
        return jdbc.sql("SELECT " + COLUMNS + " FROM rides WHERE id = :id")
                .param("id", rideId)
                .query(JdbcRideRepository::map)
                .optional();
    }

    @Override
    public Optional<Ride> findForUpdate(UUID rideId) {
        return jdbc.sql("SELECT " + COLUMNS + " FROM rides WHERE id = :id FOR UPDATE")
                .param("id", rideId)
                .query(JdbcRideRepository::map)
                .optional();
    }

    @Override
    public List<Ride> claimUndispatched(Instant requestedBefore, int limit) {
        return jdbc.sql("""
                        UPDATE rides r SET dispatch_requested_at = :now
                          FROM (SELECT id FROM rides
                                 WHERE status = 'REQUESTED' AND dispatch_requested_at < :requestedBefore
                                 ORDER BY dispatch_requested_at
                                 LIMIT :limit
                                 FOR UPDATE SKIP LOCKED) due
                         WHERE r.id = due.id
                        RETURNING\s""" + QUALIFIED_COLUMNS)
                .param("now", Timestamps.toDb(clock.instant()))
                .param("requestedBefore", Timestamps.toDb(requestedBefore))
                .param("limit", limit)
                .query(JdbcRideRepository::map)
                .list();
    }

    /**
     * Compare-and-set of the ride status, guarded by the transition's own {@code from} statuses and, where the
     * transition requires it, by the assigned driver. The CTE locks the row and exposes the status it had before the
     * update, which the history records.
     */
    @Override
    public Optional<TransitionResult> transition(UUID rideId, RideTransition transition, String driverId) {
        String assignedDriver = switch (transition.assignedDriver()) {
            case KEEP -> "";
            case SET -> ", assigned_driver_id = :driverId";
            case CLEAR -> ", assigned_driver_id = NULL";
            case KEEP_IF_ACCEPTED -> ", assigned_driver_id = CASE WHEN p.status = 'DRIVER_ACCEPTED' "
                    + "THEN r.assigned_driver_id END";
        };
        String holder = transition.requiresAssignedDriver() ? " AND r.assigned_driver_id = :driverId" : "";
        String sql = """
                WITH p AS (SELECT id, status FROM rides WHERE id = :id FOR UPDATE)
                UPDATE rides r
                   SET status = :to, version = r.version + 1, updated_at = :now""" + assignedDriver + """

                  FROM p
                 WHERE r.id = p.id AND p.status IN (:from)""" + holder + """

                RETURNING\s""" + QUALIFIED_COLUMNS + ", p.status AS previous_status";
        JdbcClient.StatementSpec statement = jdbc.sql(sql)
                .param("id", rideId)
                .param("to", transition.to().name())
                .param("from", transition.fromNames())
                .param("now", Timestamps.toDb(clock.instant()));
        if (transition.namesDriver()) {
            statement = statement.param("driverId", requireDriver(transition, driverId));
        }
        return statement
                .query((resultSet, rowNumber) -> new TransitionResult(
                        map(resultSet, rowNumber), RideStatus.valueOf(resultSet.getString("previous_status"))))
                .optional();
    }

    @Override
    public void recordHistory(UUID rideId, RideStatus from, RideStatus to, String driverId, String reason) {
        jdbc.sql("""
                        INSERT INTO ride_state_history (ride_id, from_status, to_status, driver_id, reason, occurred_at)
                        VALUES (:rideId, :fromStatus, :toStatus, :driverId, :reason, :occurredAt)
                        """)
                .param("rideId", rideId)
                .param("fromStatus", from == null ? null : from.name(), Types.VARCHAR)
                .param("toStatus", to.name())
                .param("driverId", driverId, Types.VARCHAR)
                .param("reason", reason)
                .param("occurredAt", Timestamps.toDb(clock.instant()))
                .update();
    }

    @Override
    public List<RideHistoryEntry> history(UUID rideId) {
        return jdbc.sql("""
                        SELECT from_status, to_status, driver_id, reason, occurred_at
                          FROM ride_state_history
                         WHERE ride_id = :rideId
                         ORDER BY id
                        """)
                .param("rideId", rideId)
                .query((resultSet, rowNumber) -> {
                    String from = resultSet.getString("from_status");
                    return new RideHistoryEntry(
                            from == null ? null : RideStatus.valueOf(from),
                            RideStatus.valueOf(resultSet.getString("to_status")),
                            resultSet.getString("driver_id"),
                            resultSet.getString("reason"),
                            Timestamps.fromDb(resultSet, "occurred_at"));
                })
                .list();
    }

    private static String requireDriver(RideTransition transition, String driverId) {
        if (driverId == null) {
            throw new IllegalArgumentException(transition + " needs a driver id");
        }
        return driverId;
    }

    private static Ride map(ResultSet resultSet, int rowNumber) throws SQLException {
        return new Ride(
                resultSet.getObject("id", UUID.class),
                resultSet.getString("passenger_id"),
                new Location(resultSet.getDouble("pickup_latitude"), resultSet.getDouble("pickup_longitude")),
                new Location(resultSet.getDouble("destination_latitude"), resultSet.getDouble("destination_longitude")),
                RideStatus.valueOf(resultSet.getString("status")),
                resultSet.getString("assigned_driver_id"),
                resultSet.getLong("version"),
                Timestamps.fromDb(resultSet, "created_at"),
                Timestamps.fromDb(resultSet, "updated_at"));
    }
}
