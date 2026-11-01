package com.ridehailing.driver.domain;

import com.ridehailing.common.kernel.StateTransition;

import java.util.Optional;
import java.util.Set;

import static com.ridehailing.driver.domain.DriverStatus.AVAILABLE;
import static com.ridehailing.driver.domain.DriverStatus.OFFLINE;

/**
 * What the driver may change: going online or offline. Evaluated against {@link Driver#status()}, so an
 * OFFERED or BUSY driver is rejected. Offers and trips are dispatch's assignment, not a driver transition.
 */
public enum DriverTransition implements StateTransition<DriverStatus> {
    GO_ONLINE(AVAILABLE, OFFLINE),
    GO_OFFLINE(OFFLINE, AVAILABLE);

    private final DriverStatus to;
    private final Set<DriverStatus> from;

    DriverTransition(DriverStatus to, DriverStatus from) {
        this.to = to;
        this.from = Set.of(from);
    }

    @Override
    public DriverStatus to() {
        return to;
    }

    @Override
    public Set<DriverStatus> from() {
        return from;
    }

    /**
     * The availability API may only ask for AVAILABLE or OFFLINE; empty for OFFERED and BUSY.
     */
    public static Optional<DriverTransition> toward(DriverStatus requested) {
        return switch (requested) {
            case AVAILABLE -> Optional.of(GO_ONLINE);
            case OFFLINE -> Optional.of(GO_OFFLINE);
            case OFFERED, BUSY -> Optional.empty();
        };
    }
}
