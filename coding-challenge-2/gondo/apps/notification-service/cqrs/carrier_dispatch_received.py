"""Apply Queue → Send-to-carrier from carrier ``sms.dispatch.received`` integration event."""

from __future__ import annotations

import logging
from datetime import datetime, timezone

from models import TransitionEvent, TransitionOutcome, TransitionSource, is_valid_transition
from pipeline_runtime import append_pipeline_event
from store import add_transition_event, find_by_message_id, update_notification

logger = logging.getLogger(__name__)


def apply_carrier_dispatch_received(message_id: str) -> str:
    """Idempotent: move ``Queue`` → ``Send-to-carrier`` when ``message_id`` matches.

    Returns a short result token for logging: ``applied``, ``idempotent``, ``skipped_not_queue``,
    ``not_found``.
    """
    mid = message_id.strip()
    if not mid:
        return "not_found"

    n = find_by_message_id(mid)
    if n is None:
        logger.warning("carrier_dispatch_received_unknown_message_id", extra={"message_id": mid})
        return "not_found"

    if n.state == "Send-to-carrier":
        append_pipeline_event(
            n.notification_id,
            "bus.sms_dispatch.received.idempotent",
            {"message_id": mid, "state": "Send-to-carrier"},
        )
        return "idempotent"

    if n.state != "Queue":
        logger.info(
            "carrier_dispatch_received_skip_state",
            extra={"message_id": mid, "state": n.state},
        )
        return "skipped_not_queue"

    if not is_valid_transition(n.state, "Send-to-carrier"):
        return "skipped_not_queue"

    from_st = n.state
    n.state = "Send-to-carrier"
    n.updated_at = datetime.now(timezone.utc)
    update_notification(n)
    now = datetime.now(timezone.utc)
    add_transition_event(
        TransitionEvent(
            notification_id=n.notification_id,
            from_state=from_st,
            to_state="Send-to-carrier",
            at=now,
            source="system",
            outcome="accepted",
            reason="carrier_dispatch_received",
        )
    )
    logger.info("carrier_dispatch_received_applied", extra={"message_id": mid})
    append_pipeline_event(
        n.notification_id,
        "state.Send-to-carrier",
        {"from_state": from_st, "message_id": mid, "reason": "carrier_dispatch_received"},
    )
    return "applied"
