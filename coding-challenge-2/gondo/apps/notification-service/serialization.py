"""Read-model / API serialization (query side)."""

from __future__ import annotations

from datetime import datetime, timezone

from models import Notification


def datetime_to_iso_z(dt: datetime) -> str:
    if dt.tzinfo is None:
        dt = dt.replace(tzinfo=timezone.utc)
    return dt.astimezone(timezone.utc).isoformat().replace("+00:00", "Z")


def cost_story_payload(notification: Notification) -> dict[str, object]:
    """Maps challenge brief (estimate at Send-to-provider, actual at terminal callback) to API visibility."""
    est = notification.estimated_cost is not None
    act = (
        notification.last_actual_cost is not None
        or notification.charging_actual_cost_id is not None
    )
    return {
        "estimated_cost_lifecycle_stage": "Send-to-provider",
        "estimated_available": est,
        "actual_cost_lifecycle_stage": "Send-success or Send-failed (via provider callback)",
        "actual_available": act,
        "estimate_source": "charging-service EstimateCost on dispatch/retry before entering Send-to-provider",
        "actual_source": "charging-service RecordActualCost on provider callback after terminal state update",
    }


def notification_to_dict(notification: Notification) -> dict[str, object]:
    return {
        "notification_id": notification.notification_id,
        "message_id": notification.message_id,
        "channel_type": notification.channel_type,
        "recipient": notification.recipient,
        "content": notification.content,
        "channel_payload": dict(notification.channel_payload),
        "state": notification.state,
        "attempt": notification.attempt,
        "selected_provider_id": notification.selected_provider_id,
        "routing_rule_version": notification.routing_rule_version,
        "estimated_cost": notification.estimated_cost,
        "estimated_currency": notification.estimated_currency,
        "charging_estimate_id": notification.charging_estimate_id,
        "charging_rate_id": notification.charging_rate_id,
        "last_actual_cost": notification.last_actual_cost,
        "actual_currency": notification.actual_currency,
        "charging_actual_cost_id": notification.charging_actual_cost_id,
        "cost_story": cost_story_payload(notification),
        "otp_challenge_id": notification.otp_challenge_id,
        "otp_expires_at": notification.otp_expires_at,
        "created_at": datetime_to_iso_z(notification.created_at),
        "updated_at": datetime_to_iso_z(notification.updated_at),
    }
