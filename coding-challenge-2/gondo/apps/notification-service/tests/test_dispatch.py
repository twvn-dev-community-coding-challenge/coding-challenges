"""Dispatch endpoint tests."""

from __future__ import annotations

from unittest.mock import AsyncMock, patch

from fastapi.testclient import TestClient

from generated import provider_pb2
from main import app


def test_dispatch_endpoint_exists() -> None:
    client = TestClient(app)
    response = client.post("/notifications/test-123/dispatch")
    assert response.status_code == 422


def test_create_notification_returns_data_envelope() -> None:
    client = TestClient(app)
    response = client.post(
        "/notifications",
        json={
            "message_id": "msg-dispatch-env",
            "channel_type": "SMS",
            "recipient": "u",
            "content": "hello",
            "channel_payload": {"country_code": "US", "phone_number": "+15559876543"},
        },
    )
    assert response.status_code == 201
    body = response.json()
    assert "data" in body
    assert body["data"]["message_id"] == "msg-dispatch-env"


def test_dispatch_happy_path() -> None:
    client = TestClient(app)
    created = client.post(
        "/notifications",
        json={
            "message_id": "msg-dispatch-1",
            "channel_type": "SMS",
            "recipient": "u",
            "content": "hello",
            "channel_payload": {"country_code": "US", "phone_number": "+84901234567"},
        },
    ).json()
    nid = created["data"]["notification_id"]

    mock_resp = provider_pb2.SelectProviderResponse()
    mock_resp.selected_provider_id = "prov-1"
    mock_resp.routing_rule_version = 2
    mock_resp.sms_dispatch_requested_published = True

    with (
        patch("main.select_provider", new_callable=AsyncMock, return_value=mock_resp),
        patch(
            "repository.resolve_carrier", new_callable=AsyncMock, return_value="VIETTEL"
        ),
    ):
        response = client.post(
            f"/notifications/{nid}/dispatch",
            json={"as_of": "2026-04-03T12:00:00.000Z"},
        )

    assert response.status_code == 200
    data = response.json()["data"]
    assert data["state"] == "Queue"
    assert data["attempt"] == 1
    assert data["selected_provider_id"] == "prov-1"
    assert data["routing_rule_version"] == 2
    assert data["channel_payload"]["carrier"] == "VIETTEL"


def test_dispatch_returns_503_when_dispatch_event_not_published() -> None:
    client = TestClient(app)
    created = client.post(
        "/notifications",
        json={
            "message_id": "msg-no-bus",
            "channel_type": "SMS",
            "recipient": "u",
            "content": "hello",
            "channel_payload": {"country_code": "US", "phone_number": "+84901234567"},
        },
    ).json()
    nid = created["data"]["notification_id"]

    mock_resp = provider_pb2.SelectProviderResponse()
    mock_resp.selected_provider_id = "prov-1"
    mock_resp.routing_rule_version = 2
    mock_resp.sms_dispatch_requested_published = False

    with (
        patch("main.select_provider", new_callable=AsyncMock, return_value=mock_resp),
        patch(
            "main.derive_carrier",
            new_callable=AsyncMock,
            return_value="VIETTEL",
        ),
    ):
        response = client.post(
            f"/notifications/{nid}/dispatch",
            json={"as_of": "2026-04-03T12:00:00.000Z"},
        )

    assert response.status_code == 503
    assert response.json()["error"]["code"] == "DISPATCH_REQUEST_NOT_PUBLISHED"
