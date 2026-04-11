"""Tests for in-memory rate lookup and estimate/actual recording."""

from __future__ import annotations

from datetime import datetime, timezone

from rates import (
    ActualCostRecord,
    EstimateRecord,
    estimate_cost,
    estimate_cost_batch,
    find_rate,
    record_actual_cost,
)

_AS_OF = datetime(2026, 6, 15, 12, 0, 0, tzinfo=timezone.utc)


def test_find_rate_vn_viettel_twilio() -> None:
    r = find_rate("prv_01", "VN", "VIETTEL", _AS_OF)
    assert r is not None
    assert r.rate_id == "rate_101"
    assert r.rate_version == 1
    assert r.unit_price == 0.015


def test_find_rate_unknown_returns_none() -> None:
    assert find_rate("prv_99", "VN", "VIETTEL", _AS_OF) is None


def test_estimate_cost_returns_record() -> None:
    rec = estimate_cost("msg-1", "prv_01", "VN", "VIETTEL", _AS_OF)
    assert rec is not None
    assert rec.message_id == "msg-1"
    assert rec.provider_id == "prv_01"
    assert rec.country_code == "VN"
    assert rec.carrier == "VIETTEL"
    assert rec.estimated_cost == 0.015
    assert rec.currency == "USD"
    assert rec.rate_id == "rate_101"
    assert rec.rate_version == 1
    assert rec.as_of == _AS_OF


def test_estimate_cost_no_rate_returns_none() -> None:
    assert estimate_cost("msg-x", "prv_99", "VN", "VIETTEL", _AS_OF) is None


def test_estimate_cost_batch_returns_multiple() -> None:
    rows = estimate_cost_batch(["prv_01", "prv_02"], "VN", "VIETTEL", _AS_OF)
    assert len(rows) == 2
    by_provider = {r.provider_id: r for r in rows}
    assert by_provider["prv_01"].rate_id == "rate_101"
    assert by_provider["prv_02"].rate_id == "rate_102"


def test_estimate_cost_batch_skips_no_rate() -> None:
    rows = estimate_cost_batch(["prv_01", "prv_99", "prv_02"], "VN", "VIETTEL", _AS_OF)
    assert len(rows) == 2
    assert {r.provider_id for r in rows} == {"prv_01", "prv_02"}


def test_record_actual_cost_stores_record() -> None:
    recorded = datetime(2026, 6, 20, 10, 0, 0, tzinfo=timezone.utc)
    rec = record_actual_cost(
        message_id="m1",
        provider_id="prv_01",
        provider_event_id="pe1",
        idempotency_key="idem-1",
        actual_cost=0.02,
        currency="USD",
        callback_state="DELIVERED",
        recorded_at=recorded,
    )
    assert isinstance(rec, ActualCostRecord)
    assert rec.idempotent_replay is False
    assert rec.actual_cost == 0.02
    assert rec.idempotency_key == "idem-1"


def test_record_actual_cost_idempotent() -> None:
    recorded = datetime(2026, 6, 20, 10, 0, 0, tzinfo=timezone.utc)
    first = record_actual_cost(
        message_id="m1",
        provider_id="prv_01",
        provider_event_id="pe1",
        idempotency_key="idem-same",
        actual_cost=0.02,
        currency="USD",
        callback_state="DELIVERED",
        recorded_at=recorded,
    )
    second = record_actual_cost(
        message_id="m1",
        provider_id="prv_01",
        provider_event_id="pe1",
        idempotency_key="idem-same",
        actual_cost=0.99,
        currency="USD",
        callback_state="OTHER",
        recorded_at=recorded,
    )
    assert first.actual_cost_id == second.actual_cost_id
    assert second.idempotent_replay is True
    assert second.actual_cost == first.actual_cost


def test_estimate_record_has_created_at() -> None:
    rec = estimate_cost("msg-c", "prv_01", "VN", "VIETTEL", _AS_OF)
    assert rec is not None
    assert isinstance(rec, EstimateRecord)
    assert rec.created_at.tzinfo is not None
