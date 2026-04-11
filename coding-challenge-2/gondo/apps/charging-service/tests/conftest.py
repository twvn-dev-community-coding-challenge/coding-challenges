"""Pytest fixtures for charging-service tests."""

from __future__ import annotations

import pytest

from rates import clear


@pytest.fixture(autouse=True)
def reset_rates_stores() -> None:
    clear()
    yield
    clear()
