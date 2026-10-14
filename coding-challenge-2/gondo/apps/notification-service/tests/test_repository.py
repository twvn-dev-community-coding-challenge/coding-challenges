"""Tests for notification-service PostgreSQL repository."""

from __future__ import annotations

import asyncio
from unittest.mock import AsyncMock, MagicMock

from repository import resolve_carrier


def test_resolve_carrier_unknown_when_empty() -> None:
    async def _run() -> None:
        session = AsyncMock()
        result = MagicMock()
        result.all.return_value = []
        session.execute = AsyncMock(return_value=result)
        out = await resolve_carrier(session, "")
        assert out == "UNKNOWN"

    asyncio.run(_run())


def test_resolve_carrier_returns_first_ordered_match() -> None:
    async def _run() -> None:
        session = AsyncMock()
        row_hi = MagicMock()
        row_hi._mapping = {
            "country_calling_code": "84",
            "national_destination": "98",
            "carrier_code": "VIETTEL",
            "match_priority": 50,
        }
        row_lo = MagicMock()
        row_lo._mapping = {
            "country_calling_code": "84",
            "national_destination": "90",
            "carrier_code": "MOBIFONE",
            "match_priority": 50,
        }
        result = MagicMock()
        result.all.return_value = [row_hi, row_lo]
        session.execute = AsyncMock(return_value=result)
        out = await resolve_carrier(session, "+84981234567")
        assert out == "VIETTEL"

    asyncio.run(_run())


def test_resolve_carrier_prefers_higher_match_priority() -> None:
    async def _run() -> None:
        session = AsyncMock()
        row_a = MagicMock()
        row_a._mapping = {
            "country_calling_code": "84",
            "national_destination": "9",
            "carrier_code": "LOW",
            "match_priority": 10,
        }
        row_b = MagicMock()
        row_b._mapping = {
            "country_calling_code": "84",
            "national_destination": "98",
            "carrier_code": "HIGH",
            "match_priority": 60,
        }
        result = MagicMock()
        result.all.return_value = [row_b, row_a]
        session.execute = AsyncMock(return_value=result)
        out = await resolve_carrier(session, "+84981234567")
        assert out == "HIGH"

    asyncio.run(_run())
