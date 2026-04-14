"""Subscribe to ``sms.dispatch.received`` (async CQRS integration — no sync HTTP to notification)."""

from __future__ import annotations

import json
import logging

from py_core.bus.contract import MessageBus
from py_core.bus.topics import SMS_DISPATCH_RECEIVED

from cqrs.carrier_dispatch_received import apply_carrier_dispatch_received

logger = logging.getLogger(__name__)


async def subscribe_to_dispatch_received(bus: MessageBus) -> None:
    """Register handler for carrier dispatch-received acks."""

    async def _on_raw(payload: bytes) -> None:
        try:
            data = json.loads(payload.decode("utf-8"))
        except (json.JSONDecodeError, UnicodeDecodeError):
            logger.warning(
                "sms_dispatch_received_non_json",
                extra={"raw": repr(payload[:200])},
            )
            return
        message_id = str(data.get("message_id", "")).strip()
        if not message_id:
            logger.warning("sms_dispatch_received_no_message_id")
            return
        apply_carrier_dispatch_received(message_id)

    await bus.subscribe(SMS_DISPATCH_RECEIVED, _on_raw)
