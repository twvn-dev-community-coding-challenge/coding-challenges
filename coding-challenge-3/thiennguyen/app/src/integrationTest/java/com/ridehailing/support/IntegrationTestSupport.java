package com.ridehailing.support;

import static org.assertj.core.api.Assertions.assertThat;

import com.ridehailing.common.kernel.geo.Location;
import com.ridehailing.driver.domain.DriverStatus;
import com.ridehailing.ride.adapters.web.model.RideResponse;
import com.ridehailing.ride.domain.RideStatus;
import java.time.Duration;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import org.awaitility.Awaitility;
import org.awaitility.core.ConditionFactory;
import org.junit.jupiter.api.BeforeEach;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.web.server.LocalServerPort;
import org.springframework.jdbc.core.simple.JdbcClient;
import tools.jackson.databind.ObjectMapper;

/**
 * Shared helpers. Dispatch decides offers and assignments in its own tables; the ride's status and the driver's
 * OFFERED/BUSY status follow from events, so assertions on them after a dispatch action wait with
 * {@link #eventually()}. Dispatch's own state ({@link #requestRow}, {@link #supplyRow}, {@link #offerRows}) is
 * consistent as soon as the command returns.
 */
@IntegrationTest
public abstract class IntegrationTestSupport {

    protected static final Location DESTINATION = new Location(10.8231, 106.6297);

    @LocalServerPort
    private int port;

    @Autowired
    protected JdbcClient jdbc;

    @Autowired
    protected ObjectMapper objectMapper;

    @Autowired
    protected MutableClock clock;

    protected Api api;
    protected Fixtures fixtures;
    protected SystemInvariants invariants;

    @BeforeEach
    void setUpSupport() {
        api = new Api(port, objectMapper);
        fixtures = new Fixtures(jdbc, clock);
        invariants = new SystemInvariants(jdbc);
    }

    protected static ConditionFactory eventually() {
        return Awaitility.await().atMost(Duration.ofSeconds(30)).pollInterval(Duration.ofMillis(100));
    }

    /** Dispatch invariants now, and the ride/driver copies once they have caught up. */
    protected void assertInvariants() {
        invariants.assertHold();
        eventually().untilAsserted(invariants::assertProjectionsConverged);
    }

    protected static Map<String, Object> rideRequest(String passengerId, Location pickup) {
        return Map.of(
                "passengerId", passengerId,
                "pickup", Map.of("latitude", pickup.latitude(), "longitude", pickup.longitude()),
                "destination", Map.of("latitude", DESTINATION.latitude(), "longitude", DESTINATION.longitude()));
    }

    protected RideResponse createRide(String passengerId, Location pickup) {
        Api.Response response = api.post("/rides", rideRequest(passengerId, pickup), "key-" + UUID.randomUUID());
        assertThat(response.status()).as(response.body()).isEqualTo(201);
        return response.as(RideResponse.class);
    }

    protected RideRow rideRow(UUID rideId) {
        return jdbc.sql("SELECT status, assigned_driver_id, version FROM rides WHERE id = :id")
                .param("id", rideId)
                .query((rs, n) -> new RideRow(RideStatus.valueOf(rs.getString("status")),
                        rs.getString("assigned_driver_id"), rs.getLong("version")))
                .single();
    }

    /** The driver module's view: the brief's status (OFFERED/BUSY while dispatch holds the driver). */
    protected DriverRow driverRow(String driverId) {
        return jdbc.sql("""
                        SELECT CASE WHEN assignment = 'NONE' THEN availability ELSE assignment END AS status,
                               current_ride_id
                          FROM drivers WHERE id = :id
                        """)
                .param("id", driverId)
                .query((rs, n) -> new DriverRow(DriverStatus.valueOf(rs.getString("status")),
                        rs.getObject("current_ride_id", UUID.class)))
                .single();
    }

    protected RequestRow requestRow(UUID rideId) {
        return jdbc.sql("SELECT status, driver_id FROM dispatch_requests WHERE ride_id = :id")
                .param("id", rideId)
                .query((rs, n) -> new RequestRow(rs.getString("status"), rs.getString("driver_id")))
                .single();
    }

    protected SupplyRow supplyRow(String driverId) {
        return jdbc.sql("""
                        SELECT online, assignment, current_ride_id, latitude, longitude
                          FROM driver_supply WHERE driver_id = :id
                        """)
                .param("id", driverId)
                .query((rs, n) -> new SupplyRow(rs.getBoolean("online"), rs.getString("assignment"),
                        rs.getObject("current_ride_id", UUID.class), rs.getObject("latitude", Double.class),
                        rs.getObject("longitude", Double.class)))
                .single();
    }

    protected List<OfferRow> offerRows(UUID rideId) {
        return jdbc.sql("SELECT driver_id, status FROM ride_offers WHERE ride_id = :id ORDER BY created_at, id")
                .param("id", rideId)
                .query((rs, n) -> new OfferRow(rs.getString("driver_id"), rs.getString("status")))
                .list();
    }

    protected long outboxCount(UUID rideId, String eventType) {
        return jdbc.sql("SELECT count(*) FROM outbox WHERE aggregate_id = :id AND event_type = :type")
                .param("id", rideId.toString())
                .param("type", eventType)
                .query(Long.class)
                .single();
    }

    protected void awaitRideStatus(UUID rideId, RideStatus status) {
        eventually().untilAsserted(() -> assertThat(rideRow(rideId).status()).isEqualTo(status));
    }

    protected void awaitDriverStatus(String driverId, DriverStatus status) {
        eventually().untilAsserted(() -> assertThat(driverRow(driverId).status()).isEqualTo(status));
    }

    protected void awaitOfferedTo(UUID rideId, String driverId) {
        eventually().untilAsserted(() -> {
            RideRow ride = rideRow(rideId);
            assertThat(ride.status()).isEqualTo(RideStatus.DRIVER_ASSIGNED);
            assertThat(ride.assignedDriverId()).isEqualTo(driverId);
        });
    }

    /** Accepts and waits until the ride reflects it (DRIVER_ACCEPTED). */
    protected void acceptAndAwait(UUID rideId, String driverId) {
        Api.Response response = accept(rideId, driverId);
        assertThat(response.status()).as(response.body()).isEqualTo(200);
        awaitRideStatus(rideId, RideStatus.DRIVER_ACCEPTED);
    }

    protected Api.Response accept(UUID rideId, String driverId) {
        return api.post("/rides/" + rideId + "/offers/" + driverId + "/accept");
    }

    protected Api.Response reject(UUID rideId, String driverId) {
        return api.post("/rides/" + rideId + "/offers/" + driverId + "/reject");
    }

    public record RideRow(RideStatus status, String assignedDriverId, long version) {
    }

    public record DriverRow(DriverStatus status, UUID currentRideId) {
    }

    public record RequestRow(String status, String driverId) {
    }

    public record SupplyRow(boolean online, String assignment, UUID currentRideId, Double latitude, Double longitude) {
    }

    public record OfferRow(String driverId, String status) {
    }
}
