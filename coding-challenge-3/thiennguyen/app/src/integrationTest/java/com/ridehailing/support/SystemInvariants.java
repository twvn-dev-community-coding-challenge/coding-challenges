package com.ridehailing.support;

import static org.assertj.core.api.Assertions.assertThat;

import com.ridehailing.ride.domain.RideTransition;
import com.ridehailing.ride.domain.RideStatus;
import java.util.List;
import org.springframework.jdbc.core.simple.JdbcClient;

/**
 * Global safety properties.
 *
 * <p>{@link #assertHold()} checks dispatch's own tables, where assignments are decided: these hold in every committed
 * state. {@link #assertProjectionsConverged()} checks that ride's and driver's copies agree with dispatch; those are
 * eventually consistent, so callers poll it. Each check is one statement, so it reads one snapshot.
 */
public final class SystemInvariants {

    private final JdbcClient jdbc;

    public SystemInvariants(JdbcClient jdbc) {
        this.jdbc = jdbc;
    }

    public void assertHold() {
        assertThat(ids("""
                SELECT driver_id FROM dispatch_requests
                 WHERE status IN ('OFFERED', 'ACCEPTED')
                 GROUP BY driver_id HAVING count(*) > 1
                """)).as("I1: no driver holds two active rides").isEmpty();

        assertThat(ids("""
                SELECT driver_id FROM ride_offers WHERE status = 'PENDING'
                 GROUP BY driver_id HAVING count(*) > 1
                UNION ALL
                SELECT CAST(ride_id AS text) FROM ride_offers WHERE status = 'PENDING'
                 GROUP BY ride_id HAVING count(*) > 1
                """)).as("I2: at most one pending offer per ride and per driver").isEmpty();

        assertThat(ids("""
                SELECT s.driver_id FROM driver_supply s LEFT JOIN dispatch_requests r ON r.ride_id = s.current_ride_id
                 WHERE (s.assignment = 'OFFERED' AND (r.status IS DISTINCT FROM 'OFFERED'
                                                      OR r.driver_id IS DISTINCT FROM s.driver_id))
                    OR (s.assignment = 'BUSY' AND (r.status IS DISTINCT FROM 'ACCEPTED'
                                                   OR r.driver_id IS DISTINCT FROM s.driver_id))
                UNION ALL
                SELECT CAST(r.ride_id AS text) FROM dispatch_requests r
                  LEFT JOIN driver_supply s ON s.driver_id = r.driver_id
                 WHERE (r.status = 'OFFERED' AND (s.assignment IS DISTINCT FROM 'OFFERED'
                                                  OR s.current_ride_id IS DISTINCT FROM r.ride_id))
                    OR (r.status = 'ACCEPTED' AND (s.assignment IS DISTINCT FROM 'BUSY'
                                                   OR s.current_ride_id IS DISTINCT FROM r.ride_id))
                """)).as("I3: a driver's assignment agrees with the request it holds").isEmpty();

        assertThat(ids("""
                SELECT CAST(o.id AS text) FROM ride_offers o JOIN dispatch_requests r ON r.ride_id = o.ride_id
                 WHERE o.status = 'PENDING'
                   AND (r.status <> 'OFFERED' OR r.driver_id IS DISTINCT FROM o.driver_id)
                """)).as("I4: a pending offer belongs to the request's offered driver").isEmpty();

        assertThat(ids("""
                SELECT CAST(o.id AS text) FROM ride_offers o
                  JOIN ride_driver_exclusions e ON e.ride_id = o.ride_id AND e.driver_id = o.driver_id
                 WHERE o.created_at > e.created_at
                """)).as("I5: no excluded driver is offered the same ride again").isEmpty();

        List<String> edges = jdbc.sql("SELECT DISTINCT COALESCE(from_status, '') || '>' || to_status FROM ride_state_history")
                .query(String.class)
                .list();
        assertThat(edges).as("I6: history only contains legal edges").allSatisfy(edge -> {
            String[] parts = edge.split(">");
            RideStatus to = RideStatus.valueOf(parts[1]);
            if (parts[0].isEmpty()) {
                assertThat(to).isEqualTo(RideStatus.REQUESTED);
            } else {
                RideStatus from = RideStatus.valueOf(parts[0]);
                assertThat(RideTransition.values()).as(edge)
                        .anyMatch(transition -> transition.from().contains(from) && transition.to() == to);
            }
        });
    }

    /** Ride's and driver's copies of dispatch's decisions have caught up (poll this; it is eventually true). */
    public void assertProjectionsConverged() {
        assertThat(ids("""
                SELECT CAST(r.ride_id AS text) FROM dispatch_requests r JOIN rides ride ON ride.id = r.ride_id
                 WHERE NOT (
                       (r.status = 'WAITING' AND ride.status = 'MATCHING')
                    OR (r.status = 'OFFERED' AND ride.status = 'DRIVER_ASSIGNED' AND ride.assigned_driver_id = r.driver_id)
                    OR (r.status = 'ACCEPTED' AND ride.status = 'DRIVER_ACCEPTED' AND ride.assigned_driver_id = r.driver_id)
                    OR (r.status = 'CANCELLED' AND ride.status = 'CANCELLED')
                    OR (r.status = 'COMPLETED' AND ride.status = 'COMPLETED'))
                """)).as("I7: every ride's status agrees with dispatch").isEmpty();

        assertThat(ids("""
                SELECT d.id FROM drivers d JOIN driver_supply s ON s.driver_id = d.id
                 WHERE d.assignment <> s.assignment OR d.current_ride_id IS DISTINCT FROM s.current_ride_id
                    OR (d.availability = 'AVAILABLE') <> s.online
                """)).as("I8: every driver's copy agrees with dispatch's supply").isEmpty();
    }

    private List<String> ids(String sql) {
        return jdbc.sql(sql).query(String.class).list();
    }
}
