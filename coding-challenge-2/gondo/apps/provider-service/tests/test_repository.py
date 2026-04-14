"""Integration tests for provider-service repository (requires PostgreSQL)."""

from __future__ import annotations

import asyncio
import os
from datetime import datetime, timezone

import pytest
from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker, create_async_engine

from repository import (
    fetch_active_routing_rules,
    fetch_all_providers,
    fetch_carrier_by_code,
    fetch_provider_by_code,
    resolve_carrier_from_prefix,
)

_SKIP_NO_DB = pytest.mark.skipif(
    not os.environ.get("DATABASE_URL"),
    reason="DATABASE_URL not set (optional integration tests)",
)


@_SKIP_NO_DB
def test_fetch_provider_by_code_returns_row_for_seed_provider() -> None:
    async def _run() -> None:
        url = os.environ["DATABASE_URL"]
        engine = create_async_engine(url)
        factory = async_sessionmaker(
            engine, class_=AsyncSession, expire_on_commit=False
        )
        try:
            async with factory() as session:
                row = await fetch_provider_by_code(session, "prv_01")
                assert row is not None
                m = row._mapping
                assert m["id"] == 1
                assert m["code"] == "prv_01"
                assert isinstance(m["name"], str)
        finally:
            await engine.dispose()

    asyncio.run(_run())


@_SKIP_NO_DB
def test_fetch_all_providers_non_empty_when_seeded() -> None:
    async def _run() -> None:
        url = os.environ["DATABASE_URL"]
        engine = create_async_engine(url)
        factory = async_sessionmaker(
            engine, class_=AsyncSession, expire_on_commit=False
        )
        try:
            async with factory() as session:
                rows = await fetch_all_providers(session)
                assert len(rows) >= 1
                codes = {r._mapping["code"] for r in rows}
                assert "prv_01" in codes
        finally:
            await engine.dispose()

    asyncio.run(_run())


@_SKIP_NO_DB
def test_fetch_active_routing_rules_vn_viettel_after_version_cutover() -> None:
    """As of mid-2026, VN/VIETTEL should resolve to version 2 rule (seed data)."""

    async def _run() -> None:
        url = os.environ["DATABASE_URL"]
        engine = create_async_engine(url)
        factory = async_sessionmaker(
            engine, class_=AsyncSession, expire_on_commit=False
        )
        as_of = datetime(2026, 6, 15, 12, 0, 0, tzinfo=timezone.utc)
        try:
            async with factory() as session:
                rows = await fetch_active_routing_rules(session, "VN", "VIETTEL", as_of)
                assert len(rows) >= 1
                top = rows[0]._mapping
                assert top["routing_rule_version"] == 2
                assert top["provider_id"] == 2
        finally:
            await engine.dispose()

    asyncio.run(_run())


@_SKIP_NO_DB
def test_fetch_carrier_by_code_viettel() -> None:
    async def _run() -> None:
        url = os.environ["DATABASE_URL"]
        engine = create_async_engine(url)
        factory = async_sessionmaker(
            engine, class_=AsyncSession, expire_on_commit=False
        )
        try:
            async with factory() as session:
                row = await fetch_carrier_by_code(session, "VIETTEL")
                assert row is not None
                assert row._mapping["id"] == 1
        finally:
            await engine.dispose()

    asyncio.run(_run())


@_SKIP_NO_DB
def test_resolve_carrier_from_prefix_vn_sample() -> None:
    """Prefix 84+39 is seeded for Viettel only (deterministic match)."""

    async def _run() -> None:
        url = os.environ["DATABASE_URL"]
        engine = create_async_engine(url)
        factory = async_sessionmaker(
            engine, class_=AsyncSession, expire_on_commit=False
        )
        try:
            async with factory() as session:
                out = await resolve_carrier_from_prefix(session, "+84391234567")
                assert out == "VIETTEL"
        finally:
            await engine.dispose()

    asyncio.run(_run())
