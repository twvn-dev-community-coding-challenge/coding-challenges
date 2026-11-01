package com.ridehailing.integration;

import static org.assertj.core.api.Assertions.assertThatThrownBy;

import com.ridehailing.driver.domain.DriverStatus;
import com.ridehailing.support.IntegrationTestSupport;
import com.ridehailing.support.TestCity;
import java.time.OffsetDateTime;
import java.time.ZoneOffset;
import java.util.UUID;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.dao.DataIntegrityViolationException;
import org.springframework.transaction.support.TransactionTemplate;

/**
 * Defence in depth on the tables where dispatch decides assignments: the schema rejects a double assignment even if
 * application code regresses. Each test runs in a transaction that is rolled back, so the deliberately inconsistent
 * rows never become visible to other tests.
 */
class PersistenceConstraintsIT extends IntegrationTestSupport {

    @Autowired
    private TransactionTemplate transactions;

    private final TestCity city = TestCity.next();

    @Test
    void aDriverCannotHoldTwoActiveRides() {
        inRolledBackTransaction(() -> {
            String driver = fixtures.driver(DriverStatus.AVAILABLE, city.centre());
            insertRequest("OFFERED", driver);

            assertThatThrownBy(() -> insertRequest("ACCEPTED", driver))
                    .isInstanceOf(DataIntegrityViolationException.class);
        });
    }

    @Test
    void aDriverCannotHoldTwoPendingOffers() {
        inRolledBackTransaction(() -> {
            String driver = fixtures.driver(DriverStatus.AVAILABLE, city.centre());
            insertPendingOffer(UUID.randomUUID(), driver);

            assertThatThrownBy(() -> insertPendingOffer(UUID.randomUUID(), driver))
                    .isInstanceOf(DataIntegrityViolationException.class);
        });
    }

    @Test
    void anOfferedDriverMustReferenceARide() {
        inRolledBackTransaction(() -> {
            String driver = fixtures.driver(DriverStatus.AVAILABLE, city.centre());

            assertThatThrownBy(() -> jdbc.sql("UPDATE driver_supply SET assignment = 'OFFERED' WHERE driver_id = :id")
                    .param("id", driver).update())
                    .isInstanceOf(DataIntegrityViolationException.class);
        });
    }

    @Test
    void aRideHoldsAtMostOneDriver() {
        inRolledBackTransaction(() -> {
            UUID rideId = UUID.randomUUID();
            String first = fixtures.driver(DriverStatus.AVAILABLE, city.centre());
            String second = fixtures.driver(DriverStatus.AVAILABLE, city.centre());
            assignInSupply(first, rideId);

            assertThatThrownBy(() -> assignInSupply(second, rideId))
                    .isInstanceOf(DataIntegrityViolationException.class);
        });
    }

    private void inRolledBackTransaction(Runnable work) {
        transactions.executeWithoutResult(status -> {
            status.setRollbackOnly();
            work.run();
        });
    }

    private void insertRequest(String status, String driverId) {
        OffsetDateTime now = OffsetDateTime.now(ZoneOffset.UTC);
        jdbc.sql("""
                        INSERT INTO dispatch_requests (ride_id, pickup_latitude, pickup_longitude, status, driver_id,
                                                       created_at, updated_at)
                        VALUES (:rideId, 1, 1, :status, :driver, :now, :now)
                        """)
                .param("rideId", UUID.randomUUID())
                .param("status", status)
                .param("driver", driverId)
                .param("now", now)
                .update();
    }

    private void insertPendingOffer(UUID rideId, String driverId) {
        OffsetDateTime now = OffsetDateTime.now(ZoneOffset.UTC);
        jdbc.sql("""
                        INSERT INTO ride_offers (id, ride_id, driver_id, status, distance_km, created_at, expires_at)
                        VALUES (:id, :ride, :driver, 'PENDING', 0.1, :now, :expires)
                        """)
                .param("id", UUID.randomUUID())
                .param("ride", rideId)
                .param("driver", driverId)
                .param("now", now)
                .param("expires", now.plusYears(1))
                .update();
    }

    private void assignInSupply(String driverId, UUID rideId) {
        jdbc.sql("UPDATE driver_supply SET assignment = 'OFFERED', current_ride_id = :ride WHERE driver_id = :id")
                .param("ride", rideId)
                .param("id", driverId)
                .update();
    }
}
