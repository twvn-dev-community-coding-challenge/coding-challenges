"""Tests for in-memory provider registry and routing."""

from __future__ import annotations

from datetime import datetime, timezone

from registry import (
    get_provider,
    resolve_routing,
    select_provider,
)


def _as_of() -> datetime:
    return datetime(2026, 6, 15, 12, 0, 0, tzinfo=timezone.utc)


def test_resolve_routing_vn_viettel_returns_two_candidates() -> None:
    candidates = resolve_routing("VN", "VIETTEL", _as_of())
    assert len(candidates) == 2
    rule_ids = {c.routing_rule_id for c in candidates}
    assert rule_ids == {"rr_01", "rr_02"}


def test_resolve_routing_unknown_country_returns_empty() -> None:
    assert resolve_routing("XX", "UNKNOWN", _as_of()) == []


def test_resolve_routing_orders_by_precedence_desc() -> None:
    candidates = resolve_routing("VN", "VIETTEL", _as_of())
    precedences = [c.precedence for c in candidates]
    assert precedences == sorted(precedences, reverse=True)
    assert candidates[0].precedence >= candidates[1].precedence


def test_select_provider_highest_precedence_picks_top() -> None:
    result = select_provider("VN", "VIETTEL", _as_of(), "highest_precedence")
    assert result is not None
    assert result.selected_provider_id == "prv_02"
    assert result.selected_provider_code == "VONAGE"


def test_select_provider_no_match_returns_none() -> None:
    assert select_provider("XX", "UNKNOWN", _as_of(), "highest_precedence") is None


def test_get_provider_returns_provider() -> None:
    p = get_provider("prv_01")
    assert p is not None
    assert p.provider_code == "TWILIO"


def test_get_provider_unknown_returns_none() -> None:
    assert get_provider("prv_unknown") is None
