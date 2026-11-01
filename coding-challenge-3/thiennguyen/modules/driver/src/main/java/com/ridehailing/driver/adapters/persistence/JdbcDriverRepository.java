package com.ridehailing.driver.adapters.persistence;

import com.ridehailing.common.jdbc.Timestamps;
import com.ridehailing.common.kernel.geo.Location;
import com.ridehailing.driver.application.port.DriverRepository;
import com.ridehailing.driver.domain.*;
import lombok.RequiredArgsConstructor;
import org.springframework.jdbc.core.simple.JdbcClient;
import org.springframework.stereotype.Repository;

import java.sql.ResultSet;
import java.sql.SQLException;
import java.sql.Types;
import java.time.Clock;
import java.util.Optional;
import java.util.UUID;

@Repository
@RequiredArgsConstructor
public class JdbcDriverRepository implements DriverRepository {

    private static final String COLUMNS =
            "id, availability, latitude, longitude, assignment, current_ride_id, version, updated_at";

    private final JdbcClient jdbc;
    private final Clock clock;

    @Override
    public Optional<Driver> find(String driverId) {
        return jdbc.sql("SELECT " + COLUMNS + " FROM drivers WHERE id = :id")
                .param("id", driverId)
                .query(JdbcDriverRepository::map)
                .optional();
    }

    @Override
    public Optional<Driver> findForUpdate(String driverId) {
        return jdbc.sql("SELECT " + COLUMNS + " FROM drivers WHERE id = :id FOR UPDATE")
                .param("id", driverId)
                .query(JdbcDriverRepository::map)
                .optional();
    }

    @Override
    public Optional<Driver> applyAvailability(String driverId, DriverTransition transition) {
        return jdbc.sql("""
                        UPDATE drivers
                           SET availability = :to, version = version + 1, updated_at = :now
                         WHERE id = :id AND availability IN (:from) AND assignment = 'NONE'
                        RETURNING\s""" + COLUMNS)
                .param("to", transition.to().name())
                .param("from", transition.fromNames())
                .param("now", Timestamps.toDb(clock.instant()))
                .param("id", driverId)
                .query(JdbcDriverRepository::map)
                .optional();
    }

    @Override
    public Optional<Driver> updateLocation(String driverId, Location location) {
        return jdbc.sql("""
                        UPDATE drivers
                           SET latitude = :latitude, longitude = :longitude, location_updated_at = :now,
                               version = version + 1, updated_at = :now
                         WHERE id = :id
                        RETURNING\s""" + COLUMNS)
                .param("latitude", location.latitude())
                .param("longitude", location.longitude())
                .param("now", Timestamps.toDb(clock.instant()))
                .param("id", driverId)
                .query(JdbcDriverRepository::map)
                .optional();
    }

    @Override
    public Optional<Driver> applyAssignment(String driverId, DriverAssignment assignment, long assignmentVersion) {
        return jdbc.sql("""
                        UPDATE drivers
                           SET assignment = :assignment, current_ride_id = :rideId, assignment_version = :assignmentVersion,
                               version = version + 1, updated_at = :now
                         WHERE id = :id AND assignment_version < :assignmentVersion
                        RETURNING\s""" + COLUMNS)
                .param("assignment", assignment.state().name())
                .param("rideId", assignment.rideId(), Types.OTHER)
                .param("assignmentVersion", assignmentVersion)
                .param("now", Timestamps.toDb(clock.instant()))
                .param("id", driverId)
                .query(JdbcDriverRepository::map)
                .optional();
    }

    private static Driver map(ResultSet resultSet, int rowNumber) throws SQLException {
        Double latitude = resultSet.getObject("latitude", Double.class);
        Double longitude = resultSet.getObject("longitude", Double.class);
        Assignment assignment = Assignment.valueOf(resultSet.getString("assignment"));
        return new Driver(
                resultSet.getString("id"),
                DriverStatus.valueOf(resultSet.getString("availability")),
                latitude == null || longitude == null ? null : new Location(latitude, longitude),
                DriverAssignment.of(assignment, resultSet.getObject("current_ride_id", UUID.class)),
                resultSet.getLong("version"),
                Timestamps.fromDb(resultSet, "updated_at"));
    }
}
