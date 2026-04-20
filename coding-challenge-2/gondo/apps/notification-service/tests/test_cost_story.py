"""Challenge §4 — estimate vs actual (API `cost_story` + field wiring)."""

from __future__ import annotations

from datetime import datetime, timezone

from main import cost_story_payload, notification_to_dict
from models import Notification


def test_cost_story_new_notification_no_charging_yet() -> None:
    n = Notification(
        notification_id="n1",
        message_id="m1",
        channel_type="SMS",
        recipient="r",
        content="c",
        channel_payload={"country_code": "VN", "phone_number": "+84901234567"},
        state="New",
        attempt=0,
        selected_provider_id=None,
        routing_rule_version=None,
        created_at=datetime.now(timezone.utc),
        updated_at=datetime.now(timezone.utc),
    )
    cs = cost_story_payload(n)
    assert cs["estimated_available"] is False
    assert cs["actual_available"] is False


def test_notification_to_dict_includes_cost_story() -> None:
    n = Notification(
        notification_id="n2",
        message_id="m2",
        channel_type="SMS",
        recipient="r",
        content="c",
        channel_payload={"country_code": "VN"},
        state="Queue",
        attempt=1,
        selected_provider_id="prv_01",
        routing_rule_version=2,
        created_at=datetime.now(timezone.utc),
        updated_at=datetime.now(timezone.utc),
        estimated_cost=0.02,
        estimated_currency="USD",
        charging_estimate_id="e1",
        charging_rate_id="r1",
    )
    d = notification_to_dict(n)
    assert "cost_story" in d
    assert d["cost_story"]["estimated_available"] is True
