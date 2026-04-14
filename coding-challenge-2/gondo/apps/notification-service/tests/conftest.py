"""Shared test fixtures."""

from __future__ import annotations

import pytest

from store import clear


@pytest.fixture(autouse=True)
def reset_store() -> None:
    clear()
    yield
    clear()
