"""Carrier dispatch-received → Queue → Send-to-carrier (apply helper, no NATS)."""

from __future__ import annotations

from datetime import datetime, timezone

import pytest

from cqrs.carrier_dispatch_received import apply_carrier_dispatch_received
from models import Notification
from store import clear, create_notification, get_notification


@pytest.fixture(autouse=True)
def _clean() -> None:
    clear()
    yield
    clear()


def test_apply_queue_to_send_to_carrier() -> None:
    now = datetime.now(timezone.utc)
    n = Notification(
        notification_id="nid-1",
        message_id="msg-async-1",
        channel_type="SMS",
        recipient="u",
        content="hello",
        channel_payload={"country_code": "VN", "phone_number": "+84x"},
        state="Queue",
        attempt=1,
        selected_provider_id="prv_01",
        selected_provider_code=None,
        routing_rule_version=1,
        created_at=now,
        updated_at=now,
    )
    create_notification(n)

    assert apply_carrier_dispatch_received("msg-async-1") == "applied"
    updated = get_notification("nid-1")
    assert updated is not None
    assert updated.state == "Send-to-carrier"


def test_apply_idempotent_when_already_send_to_carrier() -> None:
    now = datetime.now(timezone.utc)
    n = Notification(
        notification_id="nid-2",
        message_id="msg-async-2",
        channel_type="SMS",
        recipient="u",
        content="hello",
        channel_payload={},
        state="Send-to-carrier",
        attempt=1,
        selected_provider_id="p",
        selected_provider_code=None,
        routing_rule_version=1,
        created_at=now,
        updated_at=now,
    )
    create_notification(n)
    assert apply_carrier_dispatch_received("msg-async-2") == "idempotent"


def test_apply_skips_when_not_queue() -> None:
    now = datetime.now(timezone.utc)
    n = Notification(
        notification_id="nid-3",
        message_id="msg-async-3",
        channel_type="SMS",
        recipient="u",
        content="hello",
        channel_payload={},
        state="New",
        attempt=0,
        selected_provider_id=None,
        selected_provider_code=None,
        routing_rule_version=None,
        created_at=now,
        updated_at=now,
    )
    create_notification(n)
    assert apply_carrier_dispatch_received("msg-async-3") == "skipped_not_queue"


def test_apply_not_found() -> None:
    assert apply_carrier_dispatch_received("missing") == "not_found"
