"""HTTP API tests with service layer mocked (no Postgres required)."""

from __future__ import annotations

import uuid
from datetime import datetime, timezone
from unittest.mock import AsyncMock, patch

from fastapi.testclient import TestClient

from main import app
from service import VerifyOutcome


def test_health_via_app_openapi() -> None:
    client = TestClient(app)
    assert client.get("/health").status_code == 200


def test_issue_challenge_mocked() -> None:
    cid = uuid.uuid4()
    exp = datetime(2026, 6, 1, 12, 0, 0, tzinfo=timezone.utc)

    client = TestClient(app)
    with patch(
        "main.issue_challenge",
        AsyncMock(return_value=(cid, "123456", exp, 120)),
    ):
        r = client.post("/v1/challenges", json={"subject": "msg-x"})
    assert r.status_code == 200
    body = r.json()
    assert body["challenge_id"] == str(cid)
    assert body["code"] == "123456"
    assert body["ttl_seconds"] == 120


def test_verify_challenge_mocked() -> None:
    client = TestClient(app)
    with patch(
        "main.verify_challenge",
        AsyncMock(return_value=VerifyOutcome.SUCCESS),
    ):
        r = client.post(
            "/v1/verify",
            json={"challenge_id": str(uuid.uuid4()), "code": "123456"},
        )
    assert r.status_code == 200
    assert r.json()["status"] == "verified"
