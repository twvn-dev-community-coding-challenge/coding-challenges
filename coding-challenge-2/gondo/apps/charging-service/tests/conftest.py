"""Pytest fixtures for charging-service tests."""

from __future__ import annotations

import pytest_asyncio

from py_core.db import dispose_engine
from rates import clear


@pytest_asyncio.fixture(autouse=True)
async def reset_rates_stores_and_engine() -> None:
    clear()
    yield
    clear()
    await dispose_engine()
