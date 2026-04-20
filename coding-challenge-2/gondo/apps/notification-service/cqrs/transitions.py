"""Persisted transition audit (command side)."""

from __future__ import annotations

from datetime import datetime, timezone

from models import TransitionEvent, TransitionOutcome, TransitionSource
from store import add_transition_event


def record_transition(
    notification_id: str,
    from_state: str,
    to_state: str,
    source: TransitionSource,
    outcome: TransitionOutcome,
    reason: str,
) -> None:
    now = datetime.now(timezone.utc)
    add_transition_event(
        TransitionEvent(
            notification_id=notification_id,
            from_state=from_state,
            to_state=to_state,
            at=now,
            source=source,
            outcome=outcome,
            reason=reason,
        )
    )
