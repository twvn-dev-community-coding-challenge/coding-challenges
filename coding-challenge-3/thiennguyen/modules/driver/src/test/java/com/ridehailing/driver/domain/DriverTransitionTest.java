package com.ridehailing.driver.domain;

import static com.ridehailing.common.kernel.TransitionOutcome.ALREADY_APPLIED;
import static com.ridehailing.common.kernel.TransitionOutcome.APPLY;
import static com.ridehailing.common.kernel.TransitionOutcome.REJECTED;
import static com.ridehailing.driver.domain.DriverStatus.AVAILABLE;
import static com.ridehailing.driver.domain.DriverStatus.BUSY;
import static com.ridehailing.driver.domain.DriverStatus.OFFERED;
import static com.ridehailing.driver.domain.DriverStatus.OFFLINE;
import static org.assertj.core.api.Assertions.assertThat;

import java.time.Instant;
import java.util.UUID;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.EnumSource;

class DriverTransitionTest {

    @Test
    void aDriverSwitchesBetweenOfflineAndAvailable() {
        assertThat(DriverTransition.GO_ONLINE.evaluate(OFFLINE)).isEqualTo(APPLY);
        assertThat(DriverTransition.GO_ONLINE.to()).isEqualTo(AVAILABLE);
        assertThat(DriverTransition.GO_OFFLINE.evaluate(AVAILABLE)).isEqualTo(APPLY);
        assertThat(DriverTransition.GO_OFFLINE.to()).isEqualTo(OFFLINE);
    }

    @Test
    void repeatingAnAvailabilityChangeIsANoOp() {
        assertThat(DriverTransition.GO_ONLINE.evaluate(AVAILABLE)).isEqualTo(ALREADY_APPLIED);
        assertThat(DriverTransition.GO_OFFLINE.evaluate(OFFLINE)).isEqualTo(ALREADY_APPLIED);
    }

    @ParameterizedTest
    @EnumSource(DriverTransition.class)
    void anOfferedOrBusyDriverCannotChangeAvailability(DriverTransition transition) {
        assertThat(transition.evaluate(OFFERED)).isEqualTo(REJECTED);
        assertThat(transition.evaluate(BUSY)).isEqualTo(REJECTED);
    }

    @Test
    void theAvailabilityApiCanOnlyAskForAvailableOrOffline() {
        assertThat(DriverTransition.toward(AVAILABLE)).contains(DriverTransition.GO_ONLINE);
        assertThat(DriverTransition.toward(OFFLINE)).contains(DriverTransition.GO_OFFLINE);
        assertThat(DriverTransition.toward(OFFERED)).isEmpty();
        assertThat(DriverTransition.toward(BUSY)).isEmpty();
    }

    @Test
    void dispatchsAssignmentOverridesTheDriversChoiceInTheirStatus() {
        UUID rideId = UUID.randomUUID();

        assertThat(driver(AVAILABLE, DriverAssignment.NONE).status()).isEqualTo(AVAILABLE);
        assertThat(driver(OFFLINE, DriverAssignment.NONE).status()).isEqualTo(OFFLINE);
        assertThat(driver(AVAILABLE, DriverAssignment.of(Assignment.OFFERED, rideId)).status()).isEqualTo(OFFERED);
        assertThat(driver(AVAILABLE, DriverAssignment.of(Assignment.BUSY, rideId)).status()).isEqualTo(BUSY);
        assertThat(DriverAssignment.of(Assignment.NONE, rideId)).isEqualTo(DriverAssignment.NONE);
    }

    private static Driver driver(DriverStatus availability, DriverAssignment assignment) {
        return new Driver("D1", availability, null, assignment, 1, Instant.parse("2026-10-30T10:00:00Z"));
    }
}
