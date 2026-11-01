package com.ridehailing.ride.domain;

import static com.ridehailing.ride.domain.RideStatus.CANCELLED;
import static com.ridehailing.ride.domain.RideStatus.COMPLETED;
import static org.assertj.core.api.Assertions.assertThat;

import com.ridehailing.common.kernel.TransitionOutcome;
import java.util.Arrays;
import java.util.EnumSet;
import java.util.HashSet;
import java.util.Set;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.CsvSource;

/** The ride lifecycle (AC-06), and the rules persistence turns into the guard and effect of each compare-and-set. */
class RideTransitionTest {

    @ParameterizedTest(name = "{0} --{1}--> {2}")
    @CsvSource({
            "REQUESTED, START_MATCHING, MATCHING",
            "MATCHING, ASSIGN_DRIVER, DRIVER_ASSIGNED",
            "DRIVER_ASSIGNED, RETURN_TO_MATCHING, MATCHING",
            "DRIVER_ASSIGNED, ACCEPT, DRIVER_ACCEPTED",
            "DRIVER_ACCEPTED, COMPLETE, COMPLETED",
            "REQUESTED, CANCEL, CANCELLED",
            "MATCHING, CANCEL, CANCELLED",
            "DRIVER_ASSIGNED, CANCEL, CANCELLED",
            "DRIVER_ACCEPTED, CANCEL, CANCELLED"
    })
    void appliesValidTransitions(RideStatus current, RideTransition transition, RideStatus target) {
        assertThat(transition.evaluate(current)).isEqualTo(TransitionOutcome.APPLY);
        assertThat(transition.to()).isEqualTo(target);
    }

    @ParameterizedTest(name = "{0} --{1}--> rejected")
    @CsvSource({
            "COMPLETED, ASSIGN_DRIVER",
            "COMPLETED, CANCEL",
            "CANCELLED, COMPLETE",
            "CANCELLED, ASSIGN_DRIVER",
            "REQUESTED, ACCEPT",
            "REQUESTED, ASSIGN_DRIVER",
            "MATCHING, COMPLETE",
            "DRIVER_ASSIGNED, COMPLETE",
            "DRIVER_ACCEPTED, RETURN_TO_MATCHING"
    })
    void rejectsInvalidTransitions(RideStatus current, RideTransition transition) {
        assertThat(transition.evaluate(current)).isEqualTo(TransitionOutcome.REJECTED);
    }

    @ParameterizedTest(name = "{0} --{1}--> already applied")
    @CsvSource({
            "DRIVER_ACCEPTED, ACCEPT",
            "CANCELLED, CANCEL",
            "COMPLETED, COMPLETE",
            "MATCHING, START_MATCHING",
            "MATCHING, RETURN_TO_MATCHING"
    })
    void treatsDuplicateTransitionsAsAlreadyApplied(RideStatus current, RideTransition transition) {
        assertThat(transition.evaluate(current)).isEqualTo(TransitionOutcome.ALREADY_APPLIED);
    }

    @Test
    void legalEdgesAreExactlyTheDocumentedLifecycle() {
        Set<String> edges = new HashSet<>();
        for (RideTransition transition : RideTransition.values()) {
            transition.from().forEach(from -> edges.add(from + "->" + transition.to()));
        }

        assertThat(edges).containsExactlyInAnyOrder(
                "REQUESTED->MATCHING", "MATCHING->DRIVER_ASSIGNED", "DRIVER_ASSIGNED->MATCHING",
                "DRIVER_ASSIGNED->DRIVER_ACCEPTED", "DRIVER_ACCEPTED->COMPLETED",
                "REQUESTED->CANCELLED", "MATCHING->CANCELLED", "DRIVER_ASSIGNED->CANCELLED",
                "DRIVER_ACCEPTED->CANCELLED");
    }

    @Test
    void terminalRidesNeverApplyATransition() {
        for (RideStatus terminal : EnumSet.of(CANCELLED, COMPLETED)) {
            assertThat(terminal.isTerminal()).isTrue();
            for (RideTransition transition : RideTransition.values()) {
                assertThat(transition.evaluate(terminal)).as("%s from %s", transition, terminal)
                        .isNotEqualTo(TransitionOutcome.APPLY);
            }
        }
    }

    @Test
    void onlyAnsweringTheOfferRequiresTheAssignedDriver() {
        assertThat(Arrays.stream(RideTransition.values()).filter(RideTransition::requiresAssignedDriver))
                .containsExactlyInAnyOrder(RideTransition.RETURN_TO_MATCHING, RideTransition.ACCEPT);
    }

    @Test
    void assigningSetsTheDriverAndReturningToMatchingClearsIt() {
        assertThat(RideTransition.ASSIGN_DRIVER.assignedDriver()).isEqualTo(RideTransition.AssignedDriver.SET);
        assertThat(RideTransition.RETURN_TO_MATCHING.assignedDriver()).isEqualTo(RideTransition.AssignedDriver.CLEAR);
        assertThat(RideTransition.CANCEL.assignedDriver())
                .isEqualTo(RideTransition.AssignedDriver.KEEP_IF_ACCEPTED);
    }

    @Test
    void transitionsThatTouchADriverNameOne() {
        assertThat(Arrays.stream(RideTransition.values()).filter(RideTransition::namesDriver))
                .containsExactlyInAnyOrder(
                        RideTransition.ASSIGN_DRIVER, RideTransition.RETURN_TO_MATCHING, RideTransition.ACCEPT);
    }
}
