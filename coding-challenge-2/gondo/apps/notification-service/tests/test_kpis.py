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
    provider_code: str | None = None,
    est: float | None = None,
    actual: float | None = None,
    created_at: datetime | None = None,
    updated_at: datetime | None = None,
    calling_domain: str | None = None,
) -> Notification:
    now = datetime.now(timezone.utc)
    ca = created_at if created_at is not None else now
    ua = updated_at if updated_at is not None else ca
    cp: dict[str, str] = {"country_code": country, "phone_number": "+100"}
    if calling_domain:
        cp["calling_domain"] = calling_domain
    return Notification(
        notification_id=nid,
        message_id=mid,
        channel_type="SMS",
        recipient="r",
        content="c",
        channel_payload=cp,
        state=state,
        attempt=1,
        selected_provider_id=provider,
        selected_provider_code=provider_code,
        routing_rule_version=1,
        created_at=ca,
        updated_at=ua,
        estimated_cost=est,
        last_actual_cost=actual,
    )


def test_kpis_empty() -> None:
    data = build_sms_kpis()
    assert data["source"] == "in_memory_notifications"
    assert data["created_at_filter"]["from"] is None
    assert data["created_at_filter"]["to"] is None
    assert data["overall"]["total_notifications"] == 0
    assert data["by_provider"] == []
    assert data["by_country"] == []
    assert data["by_calling_domain"] == []


def test_kpis_groups_provider_and_country() -> None:
    create_notification(
        _n(
            nid="a",
            mid="m1",
            state="Send-success",
            provider="prv_01",
            provider_code="TWILIO",
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
            provider_code="TWILIO",
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
    assert by_p["prv_01"]["provider_code"] == "TWILIO"
    assert by_p["unassigned"]["volume"] == 1
    assert by_p["unassigned"]["provider_code"] is None

    by_c = {r["country_code"]: r for r in data["by_country"]}
    assert by_c["VN"]["volume"] == 2
    assert by_c["TH"]["volume"] == 1

    by_cd = {r["calling_domain"]: r for r in data["by_calling_domain"]}
    assert by_cd["unattributed"]["volume"] == 3


def test_kpis_groups_by_calling_domain() -> None:
    create_notification(
        _n(
            nid="cd1",
            mid="mc1",
            state="Send-success",
            provider="p",
            country="VN",
            calling_domain="booking",
        )
    )
    create_notification(
        _n(
            nid="cd2",
            mid="mc2",
            state="Send-success",
            provider="p",
            country="VN",
            calling_domain="booking",
        )
    )
    create_notification(
        _n(
            nid="cd3",
            mid="mc3",
            state="New",
            provider=None,
            country="TH",
            calling_domain="erp",
        )
    )

    data = build_sms_kpis()
    by_cd = {r["calling_domain"]: r for r in data["by_calling_domain"]}
    assert by_cd["booking"]["volume"] == 2
    assert by_cd["erp"]["volume"] == 1


def test_kpis_created_at_window() -> None:
    t0 = datetime(2025, 1, 10, 12, 0, 0, tzinfo=timezone.utc)
    t1 = datetime(2025, 1, 15, 12, 0, 0, tzinfo=timezone.utc)
    t2 = datetime(2025, 1, 20, 12, 0, 0, tzinfo=timezone.utc)
    create_notification(
        _n(
            nid="a",
            mid="m-a",
            state="Send-success",
            provider="p1",
            country="VN",
            created_at=t0,
        )
    )
    create_notification(
        _n(
            nid="b",
            mid="m-b",
            state="Send-success",
            provider="p1",
            country="VN",
            created_at=t1,
        )
    )
    create_notification(
        _n(
            nid="c",
            mid="m-c",
            state="Send-success",
            provider="p1",
            country="VN",
            created_at=t2,
        )
    )

    assert build_sms_kpis()["overall"]["total_notifications"] == 3
    assert build_sms_kpis(created_from=t1, created_to=t1)["overall"]["total_notifications"] == 1
    assert build_sms_kpis(created_from=t0, created_to=t2)["overall"]["total_notifications"] == 3
    assert build_sms_kpis(created_to=t0)["overall"]["total_notifications"] == 1
    assert build_sms_kpis(created_from=t2)["overall"]["total_notifications"] == 1


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
    assert payload["data"]["created_at_filter"]["from"] is None
    assert payload["data"]["overall"]["total_notifications"] >= 1


def test_get_kpis_http_query_window_and_validation() -> None:
    from fastapi.testclient import TestClient

    from kpis import parse_iso_datetime
    from main import app

    t_lo = datetime(2025, 6, 1, 0, 0, 0, tzinfo=timezone.utc)
    t_mid = datetime(2025, 6, 15, 0, 0, 0, tzinfo=timezone.utc)
    t_hi = datetime(2025, 7, 1, 0, 0, 0, tzinfo=timezone.utc)
    create_notification(
        _n(
            nid="w1",
            mid="mw1",
            state="Queue",
            provider=None,
            country="SG",
            created_at=t_lo,
        )
    )
    create_notification(
        _n(
            nid="w2",
            mid="mw2",
            state="Queue",
            provider=None,
            country="SG",
            created_at=t_mid,
        )
    )
    client = TestClient(app)
    r = client.get(
        "/notifications/kpis",
        params={
            "from": t_mid.isoformat(),
            "to": t_mid.isoformat(),
        },
    )
    assert r.status_code == 200
    assert r.json()["data"]["overall"]["total_notifications"] == 1

    bad = client.get("/notifications/kpis", params={"from": "not-iso"})
    assert bad.status_code == 422

    disorder = client.get(
        "/notifications/kpis",
        params={
            "from": parse_iso_datetime("2025-08-02T00:00:00Z").isoformat(),
            "to": parse_iso_datetime("2025-08-01T00:00:00Z").isoformat(),
        },
    )
    assert disorder.status_code == 422

    empty_window = client.get(
        "/notifications/kpis",
        params={
            "from": t_hi.isoformat(),
            "to": t_hi.isoformat(),
        },
    )
    assert empty_window.status_code == 200
    assert empty_window.json()["data"]["overall"]["total_notifications"] == 0
