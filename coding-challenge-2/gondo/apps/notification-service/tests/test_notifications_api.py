"""Tests for notification CRUD HTTP endpoints."""

from __future__ import annotations

from fastapi.testclient import TestClient

from main import app


def _client() -> TestClient:
    return TestClient(app)


def test_create_notification_rejects_duplicate_message_id() -> None:
    client = _client()
    body = {
        "message_id": "msg-dup-1",
        "channel_type": "SMS",
        "recipient": "a",
        "content": "hi",
        "channel_payload": {"country_code": "US", "phone_number": "+15551234567"},
    }
    assert client.post("/notifications", json=body).status_code == 201
    dup = client.post("/notifications", json=body)
    assert dup.status_code == 422
    assert dup.json()["error"]["code"] == "VALIDATION_ERROR"


def test_create_notification_success() -> None:
    client = _client()
    body = {
        "message_id": "msg-create-1",
        "channel_type": "SMS",
        "recipient": "alice",
        "content": "hi",
        "channel_payload": {"country_code": "US", "phone_number": "+15551234567"},
    }
    response = client.post("/notifications", json=body)
    assert response.status_code == 201
    payload = response.json()
    assert "data" in payload
    assert payload["data"]["message_id"] == "msg-create-1"
    assert payload["data"]["state"] == "New"
    assert "notification_id" in payload["data"]


def test_get_notification_success() -> None:
    client = _client()
    created = client.post(
        "/notifications",
        json={
            "message_id": "msg-get-1",
            "channel_type": "SMS",
            "recipient": "bob",
            "content": "yo",
            "channel_payload": {"country_code": "VN", "phone_number": "+84901234567"},
        },
    ).json()
    nid = created["data"]["notification_id"]
    response = client.get(f"/notifications/{nid}")
    assert response.status_code == 200
    assert response.json()["data"]["notification_id"] == nid


def test_get_notification_not_found() -> None:
    client = _client()
    response = client.get("/notifications/00000000-0000-0000-0000-000000000099")
    assert response.status_code == 404
    err = response.json()["error"]
    assert err["code"] == "NOT_FOUND"


def test_list_notifications_empty() -> None:
    client = _client()
    response = client.get("/notifications")
    assert response.status_code == 200
    assert response.json()["data"]["notifications"] == []


def test_list_notifications_returns_created() -> None:
    client = _client()
    client.post(
        "/notifications",
        json={
            "message_id": "msg-list-1",
            "channel_type": "SMS",
            "recipient": "c",
            "content": "x",
            "channel_payload": {"country_code": "US", "phone_number": "+15550001111"},
        },
    )
    response = client.get("/notifications")
    assert response.status_code == 200
    items = response.json()["data"]["notifications"]
    assert len(items) == 1
    assert items[0]["message_id"] == "msg-list-1"
