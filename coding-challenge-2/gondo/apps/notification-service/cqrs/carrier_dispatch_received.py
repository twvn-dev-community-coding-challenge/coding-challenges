"""Apply Queue → Send-to-carrier from carrier ``sms.dispatch.received`` integration event."""

from __future__ import annotations

import logging
from datetime import datetime, timezone

from cqrs.carrier_auto_reject import should_auto_carrier_reject
from models import TransitionEvent, is_valid_transition
from pipeline_runtime import append_pipeline_event
from store import add_transition_event, find_by_message_id, update_notification

logger = logging.getLogger(__name__)


def apply_carrier_dispatch_received(message_id: str) -> str:
    """Idempotent: move ``Queue`` → ``Send-to-carrier`` when ``message_id`` matches.

    After handoff to **Send-to-carrier**, optional simulation moves **Send-to-carrier** →
    **Carrier-rejected** when the MNO rejects the destination MSISDN (see ``carrier_auto_reject``).

    Returns a short result token for logging: ``applied``, ``applied_carrier_rejected``,
    ``idempotent``, ``skipped_not_queue``, ``not_found``.
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

    if n.state == "Carrier-rejected":
        append_pipeline_event(
            n.notification_id,
            "bus.sms_dispatch.received.idempotent",
            {"message_id": mid, "state": "Carrier-rejected"},
        )
        return "idempotent"

    if n.state != "Queue":
        logger.info(
            "carrier_dispatch_received_skip_state",
            extra={"message_id": mid, "state": n.state},
        )
        return "skipped_not_queue"

    cc = str(n.channel_payload.get("country_code") or "")
    phone = str(n.channel_payload.get("phone_number") or "")

    if not is_valid_transition(n.state, "Send-to-carrier"):
        return "skipped_not_queue"

    from_queue = n.state
    n.state = "Send-to-carrier"
    n.updated_at = datetime.now(timezone.utc)
    update_notification(n)
    now = datetime.now(timezone.utc)
    add_transition_event(
        TransitionEvent(
            notification_id=n.notification_id,
            from_state=from_queue,
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
        {
            "from_state": from_queue,
            "message_id": mid,
            "reason": "carrier_dispatch_received",
        },
    )

    # Simulated MNO rejection after carrier attempts delivery to MSISDN
    if should_auto_carrier_reject(cc, phone):
        if not is_valid_transition("Send-to-carrier", "Carrier-rejected"):
            return "applied"
        from_sc = n.state
        n.state = "Carrier-rejected"
        n.updated_at = datetime.now(timezone.utc)
        update_notification(n)
        now_cr = datetime.now(timezone.utc)
        add_transition_event(
            TransitionEvent(
                notification_id=n.notification_id,
                from_state=from_sc,
                to_state="Carrier-rejected",
                at=now_cr,
                source="system",
                outcome="accepted",
                reason="carrier_mno_reject_simulated",
            )
        )
        logger.info(
            "carrier_dispatch_received_mno_reject",
            extra={"message_id": mid, "notification_id": n.notification_id},
        )
        append_pipeline_event(
            n.notification_id,
            "state.Carrier-rejected",
            {
                "from_state": from_sc,
                "message_id": mid,
                "reason": "mno_reject_after_send_to_carrier",
            },
        )
        return "applied_carrier_rejected"

    return "applied"
