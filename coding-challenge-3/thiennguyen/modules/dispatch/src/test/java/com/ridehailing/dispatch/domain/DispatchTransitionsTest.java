package com.ridehailing.dispatch.domain;

import static com.ridehailing.common.kernel.TransitionOutcome.ALREADY_APPLIED;
import static com.ridehailing.common.kernel.TransitionOutcome.APPLY;
import static com.ridehailing.common.kernel.TransitionOutcome.REJECTED;
import static org.assertj.core.api.Assertions.assertThat;

import java.util.Arrays;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.EnumSource;

/** The three lifecycles dispatch owns: its requests, its offers and the drivers it holds. */
class DispatchTransitionsTest {

    @ParameterizedTest
    @EnumSource(value = DispatchRequestStatus.class, names = {"CANCELLED", "COMPLETED"})
    void anEndedRequestNeverChangesAgain(DispatchRequestStatus terminal) {
        assertThat(Arrays.stream(DispatchRequestTransition.values()).filter(t -> t.evaluate(terminal) == APPLY))
                .isEmpty();
    }

    @Test
    void offeringSetsTheHolderAndReturningToWaitingClearsIt() {
        assertThat(DispatchRequestTransition.OFFER.heldDriver()).isEqualTo(DispatchRequestTransition.HeldDriver.SET);
        assertThat(DispatchRequestTransition.RETURN_TO_WAITING.heldDriver())
                .isEqualTo(DispatchRequestTransition.HeldDriver.CLEAR);
        assertThat(DispatchRequestTransition.CANCEL.heldDriver()).isEqualTo(DispatchRequestTransition.HeldDriver.KEEP);
    }

    @ParameterizedTest
    @EnumSource(OfferTransition.class)
    void anOfferIsAnsweredExactlyOnce(OfferTransition transition) {
        assertThat(transition.evaluate(OfferStatus.PENDING)).isEqualTo(APPLY);
        assertThat(transition.evaluate(transition.to())).isEqualTo(ALREADY_APPLIED);
        for (OfferStatus answered : OfferStatus.values()) {
            if (answered != OfferStatus.PENDING && answered != transition.to()) {
                assertThat(transition.evaluate(answered)).as("%s from %s", transition, answered).isEqualTo(REJECTED);
            }
        }
    }

    @Test
    void onlyAFreeDriverCanBeClaimed() {
        assertThat(AssignmentTransition.CLAIM.evaluate(Assignment.NONE)).isEqualTo(APPLY);
        assertThat(AssignmentTransition.CLAIM.evaluate(Assignment.OFFERED)).isEqualTo(ALREADY_APPLIED);
        assertThat(AssignmentTransition.CLAIM.evaluate(Assignment.BUSY)).isEqualTo(REJECTED);
        assertThat(AssignmentTransition.CLAIM.claimsFreeDriver()).isTrue();
    }

    @Test
    void aTripStartsFromAnOfferAndEitherIsReleased() {
        assertThat(AssignmentTransition.START_TRIP.evaluate(Assignment.OFFERED)).isEqualTo(APPLY);
        assertThat(AssignmentTransition.START_TRIP.evaluate(Assignment.NONE)).isEqualTo(REJECTED);
        assertThat(AssignmentTransition.RELEASE.evaluate(Assignment.OFFERED)).isEqualTo(APPLY);
        assertThat(AssignmentTransition.RELEASE.evaluate(Assignment.BUSY)).isEqualTo(APPLY);
        assertThat(AssignmentTransition.RELEASE.evaluate(Assignment.NONE)).isEqualTo(ALREADY_APPLIED);
        assertThat(AssignmentTransition.START_TRIP.claimsFreeDriver()).isFalse();
        assertThat(AssignmentTransition.RELEASE.claimsFreeDriver()).isFalse();
    }
}
