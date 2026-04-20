"""MOCK_SMS_SUCCESS_PHONES auto-transition Send-to-carrier → Send-success."""

from __future__ import annotations

import asyncio
from datetime import datetime, timezone
from unittest.mock import AsyncMock, patch

import pytest

from generated import charging_pb2
from models import Notification
from pipeline_runtime import list_pipeline_events
from store import clear, create_notification, get_notification


@pytest.fixture(autouse=True)
def _clean() -> None:
    clear()
    yield
    clear()


def test_should_autocomplete_default_msisdn(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    monkeypatch.delenv("MOCK_SMS_SUCCESS_PHONES", raising=False)
    from cqrs.dev_mock_send_success import should_autocomplete_delivery_success

    assert should_autocomplete_delivery_success("+84999999999") is True
    assert should_autocomplete_delivery_success("84999999999") is True
    assert should_autocomplete_delivery_success("+84901111111") is False
    # UI / API: national 0999999999 + country VN → same canonical digits as +84999999999
    assert should_autocomplete_delivery_success("0999999999", "VN") is True
    assert should_autocomplete_delivery_success("0999999999", None) is False
    # Membership-style 090909090 + VN → +8490909090 (default mock allowlist)
    assert should_autocomplete_delivery_success("090909090", "VN") is True
    assert should_autocomplete_delivery_success("+8490909090", "VN") is True


def test_should_autocomplete_vn_national_normalized_like_frontend(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    monkeypatch.delenv("MOCK_SMS_SUCCESS_PHONES", raising=False)
    from cqrs.dev_mock_send_success import canonical_mock_msisdn_digits

    assert canonical_mock_msisdn_digits("VN", "0999999999") == "84999999999"
    assert canonical_mock_msisdn_digits("VN", "+84999999999") == "84999999999"
    assert canonical_mock_msisdn_digits("VN", "090909090") == "8490909090"
    assert canonical_mock_msisdn_digits("VN", "+8490909090") == "8490909090"


def test_should_autocomplete_disabled_when_env_empty(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("MOCK_SMS_SUCCESS_PHONES", "")
    from cqrs.dev_mock_send_success import should_autocomplete_delivery_success

    assert should_autocomplete_delivery_success("+84999999999") is False


def test_mock_applies_send_success_and_pipeline(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    monkeypatch.delenv("MOCK_SMS_SUCCESS_PHONES", raising=False)
    from cqrs.dev_mock_send_success import apply_mock_send_success_if_eligible

    now = datetime.now(timezone.utc)
    n = Notification(
        notification_id="nid-mock",
        message_id="msg-mock-ok",
        channel_type="SMS",
        recipient="u",
        content="c",
        channel_payload={"country_code": "VN", "phone_number": "+84999999999"},
        state="Send-to-carrier",
        attempt=1,
        selected_provider_id="prv_01",
        routing_rule_version=1,
        created_at=now,
        updated_at=now,
        estimated_cost=0.02,
        estimated_currency="USD",
    )
    create_notification(n)

    rec = charging_pb2.RecordActualCostResponse()
    rec.actual_cost_id = "ac-mock-1"
    rec.message_id = "msg-mock-ok"

    with patch(
        "cqrs.charging_callbacks.record_actual_cost_grpc",
        new_callable=AsyncMock,
        return_value=rec,
    ):
        asyncio.run(apply_mock_send_success_if_eligible("msg-mock-ok"))

    updated = get_notification("nid-mock")
    assert updated is not None
    assert updated.state == "Send-success"
    assert updated.charging_actual_cost_id == "ac-mock-1"
    phases = [e["phase"] for e in list_pipeline_events("nid-mock")]
    assert "state.Send-success" in phases


def test_mock_skips_non_listed_phone(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    from cqrs.dev_mock_send_success import apply_mock_send_success_if_eligible

    monkeypatch.delenv("MOCK_SMS_SUCCESS_PHONES", raising=False)
    now = datetime.now(timezone.utc)
    n = Notification(
        notification_id="nid-other",
        message_id="msg-other",
        channel_type="SMS",
        recipient="u",
        content="c",
        channel_payload={"country_code": "VN", "phone_number": "+84901111111"},
        state="Send-to-carrier",
        attempt=1,
        selected_provider_id="prv_01",
        routing_rule_version=1,
        created_at=now,
        updated_at=now,
        estimated_cost=0.02,
        estimated_currency="USD",
    )
    create_notification(n)

    asyncio.run(apply_mock_send_success_if_eligible("msg-other"))

    updated = get_notification("nid-other")
    assert updated is not None
    assert updated.state == "Send-to-carrier"
