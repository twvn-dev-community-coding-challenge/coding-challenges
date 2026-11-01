package com.ridehailing.common.kernel;

import java.util.List;
import java.util.Set;

/**
 * One legal edge of a lifecycle: the statuses it may start from and the status it leads to. The transition is the
 * single source of a lifecycle rule; persistence turns {@link #from()} into the guard of its compare-and-set.
 */
public interface StateTransition<S extends Enum<S>> {

    S to();

    Set<S> from();

    /**
     * The {@link #from()} statuses by name, the form persistence binds into {@code status IN (:from)}.
     */
    default List<String> fromNames() {
        return from().stream().map(Enum::name).sorted().toList();
    }

    /**
     * {@code APPLY} when the transition is legal from {@code current}; {@code ALREADY_APPLIED} when {@code current}
     * already is the target status (a duplicate, answered without change); {@code REJECTED} otherwise.
     */
    default TransitionOutcome evaluate(S current) {
        if (from().contains(current)) {
            return TransitionOutcome.APPLY;
        }
        if (current == to()) {
            return TransitionOutcome.ALREADY_APPLIED;
        }
        return TransitionOutcome.REJECTED;
    }
}
