"""Subscribe to ``sms.dispatch.received`` (async CQRS integration — no sync HTTP to notification)."""

from __future__ import annotations

import json
import logging
import os

from py_core.bus.contract import MessageBus
from py_core.bus.topics import SMS_DISPATCH_RECEIVED, topic_to_subject

from cqrs.carrier_dispatch_received import apply_carrier_dispatch_received
from cqrs.dev_mock_send_success import apply_mock_send_success_if_eligible

logger = logging.getLogger(__name__)


async def subscribe_to_dispatch_received(bus: MessageBus) -> None:
    """Register handler for carrier dispatch-received acks."""

    async def _on_raw(payload: bytes) -> None:
        logger.info(
            "notification_consume_begin",
            extra={
                "topic": SMS_DISPATCH_RECEIVED,
                "payload_bytes": len(payload),
            },
        )
        try:
            data = json.loads(payload.decode("utf-8"))
        except (json.JSONDecodeError, UnicodeDecodeError):
            logger.warning(
                "sms_dispatch_received_non_json",
                extra={"raw": repr(payload[:200])},
            )
            return
        message_id = str(data.get("message_id", "")).strip()
        logger.info(
            "notification_dispatch_received_parsed",
            extra={
                "message_id": message_id or None,
                "correlation_id": data.get("correlation_id"),
            },
        )
        if not message_id:
            logger.warning("sms_dispatch_received_no_message_id")
            return
        result = apply_carrier_dispatch_received(message_id)
        if result == "applied":
            await apply_mock_send_success_if_eligible(message_id)

    await bus.subscribe(SMS_DISPATCH_RECEIVED, _on_raw)
    logger.info(
        "notification_subscribed_dispatch_received",
        extra={
            "topic": SMS_DISPATCH_RECEIVED,
            "nats_subject": topic_to_subject(SMS_DISPATCH_RECEIVED),
            "nats_url": os.environ.get("NATS_URL", "nats://127.0.0.1:4222"),
        },
    )
