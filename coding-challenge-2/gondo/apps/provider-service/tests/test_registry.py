"""Tests for provider registry (async + mocked DB)."""

from __future__ import annotations

import asyncio
from contextlib import asynccontextmanager
from datetime import datetime, timezone
from typing import AsyncIterator
from unittest.mock import AsyncMock, patch
from uuid import UUID

from registry import (
    RoutingCandidate,
    get_provider,
    resolve_routing,
    select_provider,
)


def _as_of() -> datetime:
    return datetime(2026, 6, 15, 12, 0, 0, tzinfo=timezone.utc)


class _FakeRow:
    __slots__ = ("_mapping",)

    def __init__(self, mapping: dict[str, object]) -> None:
        self._mapping = mapping


@asynccontextmanager
async def _null_session() -> AsyncIterator[None]:
    yield None


def test_resolve_routing_returns_candidates_ordered_by_precedence() -> None:
    eff = datetime(2026, 1, 1, 0, 0, 0, tzinfo=timezone.utc)
    uuid_high = UUID("a0000001-0000-4000-8000-000000000001")
    uuid_low = UUID("a0000002-0000-4000-8000-000000000002")
    rule_high = _FakeRow(
        {
            "id": uuid_high,
            "country_code": "VN",
            "carrier_id": 1,
            "provider_id": 2,
            "priority": 100,
            "routing_rule_version": 1,
            "effective_from": eff,
            "effective_to": None,
        }
    )
    rule_low = _FakeRow(
        {
            "id": uuid_low,
            "country_code": "VN",
            "carrier_id": 1,
            "provider_id": 1,
            "priority": 90,
            "routing_rule_version": 1,
            "effective_from": eff,
            "effective_to": None,
        }
    )
    p_twilio = _FakeRow(
        {"id": 1, "code": "prv_01", "name": "Twilio", "created_at": eff}
    )
    p_vonage = _FakeRow(
        {"id": 2, "code": "prv_02", "name": "Vonage", "created_at": eff}
    )

    async def _run() -> None:
        with patch("registry.get_session", _null_session):
            with patch(
                "registry.fetch_active_routing_rules",
                AsyncMock(return_value=[rule_high, rule_low]),
            ):
                with patch(
                    "registry.fetch_all_providers",
                    AsyncMock(return_value=[p_twilio, p_vonage]),
                ):
                    candidates = await resolve_routing("VN", "VIETTEL", _as_of())

        assert len(candidates) == 2
        assert candidates[0].precedence >= candidates[1].precedence
        assert candidates[0].provider_id == "prv_02"
        assert {c.routing_rule_id for c in candidates} == {
            str(uuid_high),
            str(uuid_low),
        }

    asyncio.run(_run())


def test_resolve_routing_unknown_country_returns_empty() -> None:
    async def _run() -> None:
        with patch("registry.get_session", _null_session):
            with patch(
                "registry.fetch_active_routing_rules", AsyncMock(return_value=[])
            ):
                with patch("registry.fetch_all_providers", AsyncMock(return_value=[])):
                    assert await resolve_routing("XX", "UNKNOWN", _as_of()) == []

    asyncio.run(_run())


def test_select_provider_highest_precedence_picks_top() -> None:
    eff = datetime(2026, 1, 1, 0, 0, 0, tzinfo=timezone.utc)
    c_high = RoutingCandidate(
        provider_id="prv_02",
        provider_code="VONAGE",
        routing_rule_id="rr_1",
        routing_rule_version=1,
        effective_from=eff,
        effective_to=None,
        resolved_at=_as_of(),
        precedence=100,
    )
    c_low = RoutingCandidate(
        provider_id="prv_01",
        provider_code="TWILIO",
        routing_rule_id="rr_2",
        routing_rule_version=1,
        effective_from=eff,
        effective_to=None,
        resolved_at=_as_of(),
        precedence=90,
    )

    async def _run() -> None:
        with patch("registry.resolve_routing", AsyncMock(return_value=[c_high, c_low])):
            result = await select_provider(
                "VN", "VIETTEL", _as_of(), "highest_precedence"
            )
        assert result is not None
        assert result.selected_provider_id == "prv_02"
        assert result.selected_provider_code == "VONAGE"

    asyncio.run(_run())


def test_select_provider_no_match_returns_none() -> None:
    async def _run() -> None:
        with patch("registry.resolve_routing", AsyncMock(return_value=[])):
            assert (
                await select_provider("XX", "UNKNOWN", _as_of(), "highest_precedence")
                is None
            )

    asyncio.run(_run())


def test_get_provider_returns_provider() -> None:
    eff = datetime(2026, 1, 1, 0, 0, 0, tzinfo=timezone.utc)
    row = _FakeRow({"id": 1, "code": "prv_01", "name": "Twilio", "created_at": eff})

    async def _run() -> None:
        with patch("registry.get_session", _null_session):
            with patch("registry.fetch_provider_by_code", AsyncMock(return_value=row)):
                p = await get_provider("prv_01")
        assert p is not None
        assert p.provider_code == "TWILIO"

    asyncio.run(_run())


def test_get_provider_unknown_returns_none() -> None:
    async def _run() -> None:
        with patch("registry.get_session", _null_session):
            with patch("registry.fetch_provider_by_code", AsyncMock(return_value=None)):
                assert await get_provider("prv_unknown") is None

    asyncio.run(_run())
