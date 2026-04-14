"""Integration tests hit real services (Docker Compose or local Nx)."""

from __future__ import annotations

import logging
import os

import httpx
import pytest

logger = logging.getLogger("integration")


def _emit(msg: str) -> None:
    """Visible with ``pytest -s``; also goes to logging."""
    line = f"[integration] {msg}"
    print(line, flush=True)
    logger.info("%s", msg)


def _want_integration() -> bool:
    return os.environ.get("RUN_INTEGRATION_TESTS", "").strip().lower() in (
        "1",
        "true",
        "yes",
    )


def notification_base_url() -> str:
    return os.environ.get(
        "NOTIFICATION_SERVICE_URL",
        "http://127.0.0.1:8001",
    ).rstrip("/")


@pytest.fixture(scope="session")
def integration_session() -> str:
    """Skip unless RUN_INTEGRATION_TESTS=1 and notification /health is OK."""
    if not _want_integration():
        _emit("RUN_INTEGRATION_TESTS not set — skipping integration session")
        pytest.skip(
            "Set RUN_INTEGRATION_TESTS=1 and start the stack (e.g. docker compose up).",
        )
    base = notification_base_url()
    _emit(f"Probing notification health: GET {base}/health")
    try:
        r = httpx.get(f"{base}/health", timeout=3.0)
        r.raise_for_status()
    except (httpx.HTTPError, OSError) as exc:
        _emit(f"Health check failed: {exc!s}")
        pytest.skip(f"Notification service not reachable at {base}: {exc!s}")
    _emit(f"Notification service OK (HTTP {r.status_code}) — base URL: {base}")
    return base
