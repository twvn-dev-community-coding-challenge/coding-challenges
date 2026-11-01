package com.ridehailing.support;

import com.ridehailing.common.kernel.geo.Location;
import com.ridehailing.driver.domain.DriverStatus;
import java.sql.Types;
import java.time.Clock;
import java.time.Instant;
import java.time.OffsetDateTime;
import java.time.ZoneOffset;
import java.util.UUID;
import org.springframework.jdbc.core.simple.JdbcClient;

/** Inserts test-owned records directly; passengers and drivers have no registration API (out of scope). */
public final class Fixtures {

    private final JdbcClient jdbc;
    private final Clock clock;

    public Fixtures(JdbcClient jdbc, Clock clock) {
        this.jdbc = jdbc;
        this.clock = clock;
    }

    public static String uniqueSuffix() {
        return UUID.randomUUID().toString().replace("-", "").substring(0, 12);
    }

    public String passenger() {
        String id = "P-" + uniqueSuffix();
        jdbc.sql("INSERT INTO passengers (id, name) VALUES (:id, :name)")
                .param("id", id)
                .param("name", "Test passenger " + id)
                .update();
        return id;
    }

    public String driver(DriverStatus status, Location location) {
        return driver("D-" + uniqueSuffix(), status, location);
    }

    /**
     * An idle driver in the driver module and, as if its driver-events had already been consumed, in dispatch's supply
     * copy. Offers and trips are created through the API.
     */
    public String driver(String id, DriverStatus status, Location location) {
        if (status != DriverStatus.AVAILABLE && status != DriverStatus.OFFLINE) {
            throw new IllegalArgumentException("Fixtures create idle drivers only; use the API to offer or accept");
        }
        OffsetDateTime now = now();
        Double latitude = location == null ? null : location.latitude();
        Double longitude = location == null ? null : location.longitude();
        jdbc.sql("""
                        INSERT INTO drivers (id, availability, latitude, longitude, location_updated_at, updated_at)
                        VALUES (:id, :availability, :latitude, :longitude, :now, :now)
                        """)
                .param("id", id)
                .param("availability", status.name())
                .param("latitude", latitude, Types.DOUBLE)
                .param("longitude", longitude, Types.DOUBLE)
                .param("now", now)
                .update();
        jdbc.sql("""
                        INSERT INTO driver_supply (driver_id, online, latitude, longitude, updated_at)
                        VALUES (:id, :online, :latitude, :longitude, :now)
                        """)
                .param("id", id)
                .param("online", status == DriverStatus.AVAILABLE)
                .param("latitude", latitude, Types.DOUBLE)
                .param("longitude", longitude, Types.DOUBLE)
                .param("now", now)
                .update();
        return id;
    }

    /**
     * A dispatch request already WAITING for a driver, with no ride row and no event: dispatch's state as if its
     * RideRequested had been consumed (for tests that drive dispatch directly, or simulate a lost message).
     */
    public UUID waitingDispatchRequest(Location pickup, Instant updatedAt) {
        UUID rideId = UUID.randomUUID();
        OffsetDateTime timestamp = OffsetDateTime.ofInstant(updatedAt, ZoneOffset.UTC);
        jdbc.sql("""
                        INSERT INTO dispatch_requests (ride_id, pickup_latitude, pickup_longitude, status, created_at,
                                                       updated_at)
                        VALUES (:rideId, :latitude, :longitude, 'WAITING', :ts, :ts)
                        """)
                .param("rideId", rideId)
                .param("latitude", pickup.latitude())
                .param("longitude", pickup.longitude())
                .param("ts", timestamp)
                .update();
        return rideId;
    }

    /**
     * A ride in REQUESTED whose RideRequested was published at {@code requestedAt} and never reached dispatch: no
     * outbox row and no dispatch request, as if the event had been dead-lettered (ride's state only).
     */
    public UUID rideWithLostRideRequested(String passengerId, Location pickup, Location destination,
            Instant requestedAt) {
        UUID rideId = UUID.randomUUID();
        OffsetDateTime timestamp = OffsetDateTime.ofInstant(requestedAt, ZoneOffset.UTC);
        jdbc.sql("""
                        INSERT INTO rides (id, passenger_id, pickup_latitude, pickup_longitude, destination_latitude,
                                           destination_longitude, status, created_at, updated_at,
                                           dispatch_requested_at)
                        VALUES (:id, :passengerId, :pickupLatitude, :pickupLongitude, :destinationLatitude,
                                :destinationLongitude, 'REQUESTED', :ts, :ts, :ts)
                        """)
                .param("id", rideId)
                .param("passengerId", passengerId)
                .param("pickupLatitude", pickup.latitude())
                .param("pickupLongitude", pickup.longitude())
                .param("destinationLatitude", destination.latitude())
                .param("destinationLongitude", destination.longitude())
                .param("ts", timestamp)
                .update();
        jdbc.sql("""
                        INSERT INTO ride_state_history (ride_id, from_status, to_status, reason, occurred_at)
                        VALUES (:id, NULL, 'REQUESTED', 'ride requested', :ts)
                        """)
                .param("id", rideId)
                .param("ts", timestamp)
                .update();
        return rideId;
    }

    private OffsetDateTime now() {
        return OffsetDateTime.ofInstant(clock.instant(), ZoneOffset.UTC);
    }
}
