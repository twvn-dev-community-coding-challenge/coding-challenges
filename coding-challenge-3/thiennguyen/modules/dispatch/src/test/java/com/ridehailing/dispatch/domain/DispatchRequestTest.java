package com.ridehailing.dispatch.domain;

import static com.ridehailing.common.kernel.TransitionOutcome.ALREADY_APPLIED;
import static com.ridehailing.common.kernel.TransitionOutcome.APPLY;
import static com.ridehailing.common.kernel.TransitionOutcome.REJECTED;
import static org.assertj.core.api.Assertions.assertThat;

import com.ridehailing.common.kernel.geo.Location;
import java.time.Instant;
import java.util.UUID;
import org.junit.jupiter.api.Test;

class DispatchRequestTest {

    private static final Location PICKUP = new Location(10.7769, 106.7009);

    @Test
    void aWaitingRequestCanBeOfferedToAnyDriver() {
        assertThat(request(DispatchRequestStatus.WAITING, null).evaluate(DispatchRequestTransition.OFFER, "D1"))
                .isEqualTo(APPLY);
    }

    @Test
    void onlyTheHolderMayAcceptRejectOrLetTheOfferExpire() {
        DispatchRequest offered = request(DispatchRequestStatus.OFFERED, "D1");

        assertThat(offered.evaluate(DispatchRequestTransition.ACCEPT, "D1")).isEqualTo(APPLY);
        assertThat(offered.evaluate(DispatchRequestTransition.RETURN_TO_WAITING, "D1")).isEqualTo(APPLY);
        assertThat(offered.evaluate(DispatchRequestTransition.ACCEPT, "D2")).isEqualTo(REJECTED);
        assertThat(offered.evaluate(DispatchRequestTransition.RETURN_TO_WAITING, "D2")).isEqualTo(REJECTED);
        assertThat(offered.evaluate(DispatchRequestTransition.ACCEPT, null)).isEqualTo(REJECTED);
    }

    @Test
    void aRepeatedAcceptanceByTheHolderIsAlreadyApplied() {
        DispatchRequest accepted = request(DispatchRequestStatus.ACCEPTED, "D1");

        assertThat(accepted.evaluate(DispatchRequestTransition.ACCEPT, "D1")).isEqualTo(ALREADY_APPLIED);
        assertThat(accepted.evaluate(DispatchRequestTransition.ACCEPT, "D2")).isEqualTo(REJECTED);
    }

    @Test
    void aWaitingRequestHasNoHolderSoNobodyCanAnswerIt() {
        DispatchRequest waiting = request(DispatchRequestStatus.WAITING, null);

        assertThat(waiting.evaluate(DispatchRequestTransition.ACCEPT, "D1")).isEqualTo(REJECTED);
        assertThat(waiting.evaluate(DispatchRequestTransition.RETURN_TO_WAITING, "D1")).isEqualTo(REJECTED);
    }

    @Test
    void theRideEndingAppliesWhateverTheRequestIsDoingAndOnlyOnce() {
        for (DispatchRequestStatus active : new DispatchRequestStatus[] {
                DispatchRequestStatus.WAITING, DispatchRequestStatus.OFFERED, DispatchRequestStatus.ACCEPTED}) {
            DispatchRequest request = request(active, active == DispatchRequestStatus.WAITING ? null : "D1");
            assertThat(request.evaluate(DispatchRequestTransition.CANCEL, null)).as("cancel %s", active).isEqualTo(APPLY);
            assertThat(request.evaluate(DispatchRequestTransition.COMPLETE, null)).as("complete %s", active)
                    .isEqualTo(APPLY);
        }
        DispatchRequest cancelled = request(DispatchRequestStatus.CANCELLED, "D1");
        assertThat(cancelled.evaluate(DispatchRequestTransition.CANCEL, null)).isEqualTo(ALREADY_APPLIED);
        assertThat(cancelled.evaluate(DispatchRequestTransition.COMPLETE, null)).isEqualTo(REJECTED);
        assertThat(cancelled.evaluate(DispatchRequestTransition.ACCEPT, "D1")).isEqualTo(REJECTED);
    }

    private static DispatchRequest request(DispatchRequestStatus status, String driverId) {
        Instant now = Instant.parse("2026-10-29T10:00:00Z");
        return new DispatchRequest(UUID.randomUUID(), PICKUP, status, driverId, 1, now, now);
    }
}
