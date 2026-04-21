"""OTP integration policy: issue failure and plaintext exposure flags."""

from __future__ import annotations

from unittest.mock import AsyncMock, patch

from fastapi.testclient import TestClient

from main import app
from otp_http_client import OtpIssueError


def _client() -> TestClient:
    return TestClient(app)


def test_create_returns_503_when_server_otp_issue_fails() -> None:
    client = _client()
    with patch(
        "main.issue_challenge_http",
        AsyncMock(side_effect=OtpIssueError("otp_service_unreachable")),
    ):
        response = client.post(
            "/notifications",
            json={
                "message_id": "msg-otp-down",
                "channel_type": "SMS",
                "recipient": "u",
                "content": "Code {{OTP}}",
                "channel_payload": {"country_code": "VN", "phone_number": "+84901234567"},
                "issue_server_otp": True,
            },
        )
    assert response.status_code == 503
    assert response.json()["error"]["code"] == "OTP_ISSUE_FAILED"


def test_create_server_otp_omits_plaintext_when_expose_disabled(monkeypatch) -> None:
    monkeypatch.setenv("OTP_EXPOSE_PLAINTEXT_TO_CLIENT", "false")
    client = _client()
    otp_body = {
        "challenge_id": "550e8400-e29b-41d4-a716-446655440000",
        "code": "987654",
        "expires_at": "2026-06-01T12:00:00Z",
        "ttl_seconds": 120,
    }
    with patch("main.issue_challenge_http", AsyncMock(return_value=otp_body)):
        response = client.post(
            "/notifications",
            json={
                "message_id": "msg-otp-private",
                "channel_type": "SMS",
                "recipient": "u",
                "content": "Your code {{OTP}}",
                "channel_payload": {"country_code": "VN", "phone_number": "+84901234567"},
                "issue_server_otp": True,
            },
        )
    assert response.status_code == 201
    data = response.json()["data"]
    assert "otp_plaintext" not in data
    assert data["content"] == "Your code 987654"
    assert data["otp_challenge_id"] == otp_body["challenge_id"]
