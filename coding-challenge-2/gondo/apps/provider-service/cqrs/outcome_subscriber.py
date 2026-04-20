"""Subscribe to carrier outcomes (CQRS read-side / integration)."""

from __future__ import annotations

import json
import logging

from py_core.bus.contract import MessageBus
from py_core.bus.topics import SMS_DISPATCH_OUTCOME

logger = logging.getLogger(__name__)


async def subscribe_to_dispatch_outcomes(bus: MessageBus) -> None:
    """Register handler for ``sms.dispatch.outcome`` (fire-and-forget subscription)."""

    async def _on_raw(payload: bytes) -> None:
        logger.info(
            "provider_consume_begin",
            extra={
                "topic": SMS_DISPATCH_OUTCOME,
                "payload_bytes": len(payload),
            },
        )
        try:
            data = json.loads(payload.decode("utf-8"))
        except (json.JSONDecodeError, UnicodeDecodeError):
            logger.warning(
                "sms_outcome_non_json",
                extra={"raw": repr(payload[:200])},
            )
            return
        logger.info(
            "sms_dispatch_outcome_received",
            extra={
                "status": data.get("status"),
                "message_id": data.get("message_id"),
                "correlation_id": data.get("correlation_id"),
            },
        )

    await bus.subscribe(SMS_DISPATCH_OUTCOME, _on_raw)
