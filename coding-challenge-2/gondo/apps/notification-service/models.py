"""Domain models and notification state machine."""

from __future__ import annotations

from dataclasses import dataclass
from datetime import datetime
from typing import Literal

TransitionSource = Literal["system", "callback", "retry"]
TransitionOutcome = Literal["accepted", "rejected"]

VALID_TRANSITIONS: dict[str, list[str]] = {
    "New": ["Send-to-provider"],
    "Send-to-provider": ["Queue"],
    "Queue": ["Send-to-carrier", "Carrier-rejected"],
    "Send-to-carrier": ["Send-success", "Send-failed"],
    "Send-failed": ["Send-to-provider"],
    "Carrier-rejected": ["Send-to-provider"],
}


@dataclass
class Notification:
    """In-memory notification aggregate root."""

    notification_id: str
    message_id: str
    channel_type: str
    recipient: str
    content: str
    channel_payload: dict[str, str]
    state: str
    attempt: int
    selected_provider_id: str | None
    routing_rule_version: int | None
    created_at: datetime
    updated_at: datetime


@dataclass
class TransitionEvent:
    """Recorded state transition for audit."""

    notification_id: str
    from_state: str
    to_state: str
    at: datetime
    source: TransitionSource
    outcome: TransitionOutcome
    reason: str


def is_valid_transition(from_state: str, to_state: str) -> bool:
    """Return True if transitioning from *from_state* to *to_state* is allowed."""
    allowed = VALID_TRANSITIONS.get(from_state)
    if allowed is None:
        return False
    return to_state in allowed
