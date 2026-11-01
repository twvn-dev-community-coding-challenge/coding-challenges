package com.ridehailing.dispatch.adapters.persistence;

import com.ridehailing.common.jdbc.Timestamps;
import com.ridehailing.common.kernel.geo.BoundingBox;
import com.ridehailing.common.kernel.geo.Location;
import com.ridehailing.dispatch.application.port.DriverSupplyRepository;
import com.ridehailing.dispatch.domain.Assignment;
import com.ridehailing.dispatch.domain.AssignmentTransition;
import com.ridehailing.dispatch.domain.SupplyChange;
import com.ridehailing.dispatch.domain.SupplyDriver;
import lombok.RequiredArgsConstructor;
import org.springframework.jdbc.core.simple.JdbcClient;
import org.springframework.stereotype.Repository;

import java.sql.ResultSet;
import java.sql.SQLException;
import java.time.Clock;
import java.time.OffsetDateTime;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

/**
 * Dispatch's supply table. Online flag and location are a copy of driver-events, each guarded by its own version; the
 * assignment is owned here and changes only by compare-and-set.
 */
@Repository
@RequiredArgsConstructor
public class JdbcDriverSupplyRepository implements DriverSupplyRepository {

    private static final String COLUMNS = "driver_id, online, latitude, longitude, availability_version, assignment, "
            + "current_ride_id, version";
    private static final String CHANGE_COLUMNS = "driver_id, current_ride_id, assignment, version";

    private final JdbcClient jdbc;
    private final Clock clock;

    @Override
    public Optional<SupplyDriver> find(String driverId) {
        return jdbc.sql("SELECT " + COLUMNS + " FROM driver_supply WHERE driver_id = :id")
                .param("id", driverId)
                .query(JdbcDriverSupplyRepository::map)
                .optional();
    }

    @Override
    public List<SupplyDriver> findFreeWithin(BoundingBox box) {
        return jdbc.sql("SELECT " + COLUMNS + """
                         FROM driver_supply
                        WHERE online AND assignment = 'NONE'
                          AND latitude BETWEEN :minLatitude AND :maxLatitude
                          AND longitude BETWEEN :minLongitude AND :maxLongitude
                        """)
                .param("minLatitude", box.minLatitude())
                .param("maxLatitude", box.maxLatitude())
                .param("minLongitude", box.minLongitude())
                .param("maxLongitude", box.maxLongitude())
                .query(JdbcDriverSupplyRepository::map)
                .list();
    }

    @Override
    public Optional<SupplyChange> apply(String driverId, UUID rideId, AssignmentTransition transition) {
        String guard = transition.claimsFreeDriver()
                ? " AND online AND latitude IS NOT NULL AND current_ride_id IS NULL"
                : " AND current_ride_id = :rideId";
        String rideColumn = transition.to() == Assignment.NONE ? "NULL" : ":rideId";
        String sql = "UPDATE driver_supply SET assignment = :to, current_ride_id = " + rideColumn
                + ", version = version + 1, updated_at = :now WHERE driver_id = :id AND assignment IN (:from)" + guard
                + " RETURNING " + CHANGE_COLUMNS;
        return jdbc.sql(sql)
                .param("to", transition.to().name())
                .param("now", now())
                .param("id", driverId)
                .param("rideId", rideId)
                .param("from", transition.fromNames())
                .query((resultSet, rowNumber) -> new SupplyChange(
                        resultSet.getString("driver_id"),
                        resultSet.getObject("current_ride_id", UUID.class),
                        Assignment.valueOf(resultSet.getString("assignment")),
                        resultSet.getLong("version")))
                .optional();
    }

    @Override
    public void applyAvailability(String driverId, boolean online, long availabilityVersion) {
        jdbc.sql("""
                        INSERT INTO driver_supply (driver_id, online, availability_version, updated_at)
                        VALUES (:id, :online, :version, :now)
                        ON CONFLICT (driver_id) DO UPDATE
                           SET online = EXCLUDED.online, availability_version = EXCLUDED.availability_version,
                               updated_at = EXCLUDED.updated_at
                         WHERE driver_supply.availability_version < EXCLUDED.availability_version
                        """)
                .param("id", driverId)
                .param("online", online)
                .param("version", availabilityVersion)
                .param("now", now())
                .update();
    }

    @Override
    public void applyLocation(String driverId, Location location, long locationVersion) {
        jdbc.sql("""
                        INSERT INTO driver_supply (driver_id, latitude, longitude, location_version, updated_at)
                        VALUES (:id, :latitude, :longitude, :version, :now)
                        ON CONFLICT (driver_id) DO UPDATE
                           SET latitude = EXCLUDED.latitude, longitude = EXCLUDED.longitude,
                               location_version = EXCLUDED.location_version, updated_at = EXCLUDED.updated_at
                         WHERE driver_supply.location_version < EXCLUDED.location_version
                        """)
                .param("id", driverId)
                .param("latitude", location.latitude())
                .param("longitude", location.longitude())
                .param("version", locationVersion)
                .param("now", now())
                .update();
    }

    private OffsetDateTime now() {
        return Timestamps.toDb(clock.instant());
    }

    private static SupplyDriver map(ResultSet resultSet, int rowNumber) throws SQLException {
        Double latitude = resultSet.getObject("latitude", Double.class);
        Double longitude = resultSet.getObject("longitude", Double.class);
        return new SupplyDriver(
                resultSet.getString("driver_id"),
                resultSet.getBoolean("online"),
                latitude == null || longitude == null ? null : new Location(latitude, longitude),
                resultSet.getLong("availability_version"),
                Assignment.valueOf(resultSet.getString("assignment")),
                resultSet.getObject("current_ride_id", UUID.class),
                resultSet.getLong("version"));
    }
}
