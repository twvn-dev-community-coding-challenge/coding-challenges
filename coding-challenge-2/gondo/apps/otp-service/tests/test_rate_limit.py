"""Per-IP rate limits for issue and verify (env-driven)."""

from __future__ import annotations

import uuid
from datetime import datetime, timezone
from unittest.mock import AsyncMock, patch

import pytest
from fastapi.testclient import TestClient

from main import app
from rate_limit import reset_for_testing
from service import VerifyOutcome


@pytest.fixture(autouse=True)
def _reset_rate_limit_state() -> None:
    reset_for_testing()
    yield
    reset_for_testing()


def test_issue_returns_429_when_rate_exceeded(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("OTP_ISSUE_REQUESTS_PER_MINUTE", "2")
    cid = uuid.uuid4()
    exp = datetime(2026, 6, 1, 12, 0, 0, tzinfo=timezone.utc)
    client = TestClient(app)
    with patch(
        "main.issue_challenge",
        AsyncMock(return_value=(cid, "123456", exp, 120)),
    ):
        assert client.post("/v1/challenges", json={"subject": "a"}).status_code == 200
        assert client.post("/v1/challenges", json={"subject": "b"}).status_code == 200
        r = client.post("/v1/challenges", json={"subject": "c"})
    assert r.status_code == 429
    assert r.json()["detail"] == "issue_rate_limit_exceeded"


def test_verify_returns_429_when_rate_exceeded(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("OTP_VERIFY_REQUESTS_PER_MINUTE", "2")
    cid = uuid.uuid4()
    client = TestClient(app)
    with patch(
        "main.verify_challenge",
        AsyncMock(return_value=VerifyOutcome.SUCCESS),
    ):
        body = {"challenge_id": str(cid), "code": "123456"}
        assert client.post("/v1/verify", json=body).status_code == 200
        assert client.post("/v1/verify", json=body).status_code == 200
        r = client.post("/v1/verify", json=body)
    assert r.status_code == 429
    assert r.json()["detail"] == "verify_rate_limit_exceeded"
