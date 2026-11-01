package com.ridehailing.dispatch.application;

import io.micrometer.core.instrument.MeterRegistry;
import java.time.Duration;
import java.util.Locale;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Component;
import org.springframework.transaction.support.TransactionSynchronization;
import org.springframework.transaction.support.TransactionSynchronizationManager;

/**
 * Dispatch's meters. Each is recorded once the transaction that decided it commits, so a rolled-back attempt is never
 * counted.
 * <ul>
 *   <li>{@code ridehailing.dispatch.attempts{outcome=offered|no_driver}}: matching attempts and what they found.</li>
 *   <li>{@code ridehailing.dispatch.claims.lost}: candidates another dispatch claimed first, the contention on the
 *       nearest drivers.</li>
 *   <li>{@code ridehailing.dispatch.time.to.offer{offer=first|rematch}}: from dispatch opening the request to each
 *       offer; a rematch follows a rejected or expired offer.</li>
 *   <li>{@code ridehailing.offers.ended{result}}: how offers end.</li>
 * </ul>
 */
@Component
@RequiredArgsConstructor
public class DispatchMetrics {

    public enum OfferEnd {
        ACCEPTED, REJECTED, TIMED_OUT, DRIVER_OFFLINE, CANCELLED
    }

    private final MeterRegistry meters;

    void attempt(DispatchOutcome outcome) {
        afterCommit(() -> meters.counter("ridehailing.dispatch.attempts", "outcome", tag(outcome)).increment());
    }

    void claimLost() {
        afterCommit(() -> meters.counter("ridehailing.dispatch.claims.lost").increment());
    }

    void offered(Duration sinceOpened, boolean rematch) {
        afterCommit(() -> meters.timer("ridehailing.dispatch.time.to.offer", "offer", rematch ? "rematch" : "first")
                .record(sinceOpened));
    }

    void offerEnded(OfferEnd end) {
        afterCommit(() -> meters.counter("ridehailing.offers.ended", "result", tag(end)).increment());
    }

    private static String tag(Enum<?> value) {
        return value.name().toLowerCase(Locale.ROOT);
    }

    private static void afterCommit(Runnable record) {
        if (!TransactionSynchronizationManager.isSynchronizationActive()) {
            record.run();
            return;
        }
        TransactionSynchronizationManager.registerSynchronization(new TransactionSynchronization() {
            @Override
            public void afterCommit() {
                record.run();
            }
        });
    }
}
