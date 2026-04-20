"""Charging gRPC side effects after provider terminal callbacks."""

from __future__ import annotations

import logging
from datetime import datetime, timezone

import grpc

from charging_client import record_actual_cost_grpc
from models import Notification
from pipeline_runtime import pipe
from schemas import ProviderCallbackRequest
from store import update_notification

logger = logging.getLogger(__name__)


async def record_actual_cost_after_callback(
    n: Notification,
    body: ProviderCallbackRequest,
) -> None:
    """Best-effort ``RecordActualCost`` for terminal billing (does not fail the HTTP callback)."""
    if n.selected_provider_id is None:
        return
    if body.new_state not in ("Send-success", "Send-failed"):
        return
    if body.new_state == "Send-failed" and body.actual_cost is None:
        return

    cost = body.actual_cost
    if cost is None:
        cost = n.estimated_cost
    if cost is None:
        cost = 0.0
    currency = n.estimated_currency or "USD"
    idempotency_key = f"{n.message_id}:actual:{body.new_state}"
    try:
        rec = await record_actual_cost_grpc(
            message_id=n.message_id,
            provider_id=n.selected_provider_id,
            actual_cost=float(cost),
            currency=currency,
            callback_state=body.new_state,
            idempotency_key=idempotency_key,
        )
        n.last_actual_cost = float(cost)
        n.actual_currency = currency
        n.charging_actual_cost_id = rec.actual_cost_id
        n.updated_at = datetime.now(timezone.utc)
        update_notification(n)
        pipe(
            n.notification_id,
            "charging.RecordActualCost.ok",
            message_id=n.message_id,
            charging_actual_cost_id=n.charging_actual_cost_id,
            last_actual_cost=n.last_actual_cost,
            currency=n.actual_currency,
            callback_state=body.new_state,
        )
    except grpc.aio.AioRpcError:
        logger.exception(
            "charging_record_actual_failed",
            extra={"message_id": n.message_id},
        )
