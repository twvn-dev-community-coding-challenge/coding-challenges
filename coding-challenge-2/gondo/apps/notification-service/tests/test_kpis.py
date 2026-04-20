"""GET /notifications/kpis — in-memory aggregates (User Story 5)."""

from __future__ import annotations

from datetime import datetime, timezone

import pytest

from kpis import build_sms_kpis
from models import Notification
from store import clear, create_notification


@pytest.fixture(autouse=True)
def _clean() -> None:
    clear()
    yield
    clear()


def _n(
    *,
    nid: str,
    mid: str,
    state: str,
    provider: str | None,
    country: str,
    est: float | None = None,
    actual: float | None = None,
) -> Notification:
    now = datetime.now(timezone.utc)
    return Notification(
        notification_id=nid,
        message_id=mid,
        channel_type="SMS",
        recipient="r",
        content="c",
        channel_payload={"country_code": country, "phone_number": "+100"},
        state=state,
        attempt=1,
        selected_provider_id=provider,
        routing_rule_version=1,
        created_at=now,
        updated_at=now,
        estimated_cost=est,
        last_actual_cost=actual,
    )


def test_kpis_empty() -> None:
    data = build_sms_kpis()
    assert data["source"] == "in_memory_notifications"
    assert data["overall"]["total_notifications"] == 0
    assert data["by_provider"] == []
    assert data["by_country"] == []


def test_kpis_groups_provider_and_country() -> None:
    create_notification(
        _n(
            nid="a",
            mid="m1",
            state="Send-success",
            provider="prv_01",
            country="VN",
            est=0.02,
            actual=0.019,
        )
    )
    create_notification(
        _n(
            nid="b",
            mid="m2",
            state="Send-failed",
            provider="prv_01",
            country="TH",
            est=0.03,
            actual=None,
        )
    )
    create_notification(
        _n(
            nid="c",
            mid="m3",
            state="New",
            provider=None,
            country="VN",
            est=None,
            actual=None,
        )
    )

    data = build_sms_kpis()
    assert data["overall"]["total_notifications"] == 3
    assert data["overall"]["send_success"] == 1
    assert data["overall"]["terminal_failure"] == 1
    assert data["overall"]["in_flight"] == 1

    by_p = {r["provider_id"]: r for r in data["by_provider"]}
    assert by_p["prv_01"]["volume"] == 2
    assert by_p["prv_01"]["total_estimated_cost"] == pytest.approx(0.05)
    assert by_p["unassigned"]["volume"] == 1

    by_c = {r["country_code"]: r for r in data["by_country"]}
    assert by_c["VN"]["volume"] == 2
    assert by_c["TH"]["volume"] == 1


def test_kpis_carrier_rejected_counts_as_terminal_failure() -> None:
    create_notification(
        _n(
            nid="x",
            mid="mx",
            state="Carrier-rejected",
            provider="prv_02",
            country="VN",
            est=0.01,
            actual=None,
        )
    )
    data = build_sms_kpis()
    assert data["overall"]["carrier_rejected"] == 1
    assert data["overall"]["terminal_success_rate"] == 0.0
    assert data["overall"]["terminal_failure_rate"] == 1.0


def test_get_kpis_http_endpoint() -> None:
    from fastapi.testclient import TestClient

    from main import app

    create_notification(
        _n(
            nid="h",
            mid="mh",
            state="Queue",
            provider=None,
            country="SG",
            est=0.05,
            actual=None,
        )
    )
    client = TestClient(app)
    r = client.get("/notifications/kpis")
    assert r.status_code == 200
    payload = r.json()
    assert payload["data"]["source"] == "in_memory_notifications"
    assert payload["data"]["overall"]["total_notifications"] >= 1
