"""Tests for in-memory notification store."""

from __future__ import annotations

from datetime import datetime, timezone

from models import Notification, TransitionEvent
from store import (
    add_transition_event,
    create_notification,
    find_by_message_id,
    get_notification,
    get_transition_events,
    list_notifications,
    update_notification,
)


def _sample_notification(
    notification_id: str = "nid-1",
    message_id: str = "msg-1",
) -> Notification:
    now = datetime(2026, 4, 3, 12, 0, 0, tzinfo=timezone.utc)
    return Notification(
        notification_id=notification_id,
        message_id=message_id,
        channel_type="SMS",
        recipient="u1",
        content="hello",
        channel_payload={"country_code": "US", "phone_number": "+15551234567"},
        state="New",
        attempt=0,
        selected_provider_id=None,
        routing_rule_version=None,
        created_at=now,
        updated_at=now,
    )


def test_create_and_get_notification() -> None:
    n = _sample_notification()
    create_notification(n)
    got = get_notification("nid-1")
    assert got is not None
    assert got.message_id == "msg-1"


def test_find_by_message_id() -> None:
    create_notification(_sample_notification())
    found = find_by_message_id("msg-1")
    assert found is not None
    assert found.notification_id == "nid-1"


def test_list_notifications() -> None:
    create_notification(_sample_notification(notification_id="a", message_id="m-a"))
    create_notification(_sample_notification(notification_id="b", message_id="m-b"))
    items = list_notifications()
    assert len(items) == 2


def test_update_notification() -> None:
    n = _sample_notification()
    create_notification(n)
    n.state = "Queue"
    n.updated_at = datetime(2026, 4, 4, 12, 0, 0, tzinfo=timezone.utc)
    update_notification(n)
    got = get_notification("nid-1")
    assert got is not None
    assert got.state == "Queue"


def test_get_nonexistent_returns_none() -> None:
    assert get_notification("missing") is None


def test_add_and_get_transition_events() -> None:
    create_notification(_sample_notification())
    evt = TransitionEvent(
        notification_id="nid-1",
        from_state="New",
        to_state="Send-to-provider",
        at=datetime(2026, 4, 3, 12, 0, 0, tzinfo=timezone.utc),
        source="system",
        outcome="accepted",
        reason="dispatch",
    )
    add_transition_event(evt)
    events = get_transition_events("nid-1")
    assert len(events) == 1
    assert events[0].to_state == "Send-to-provider"
