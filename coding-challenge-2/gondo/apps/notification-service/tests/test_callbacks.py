"""Tests for provider callback ingestion."""

from __future__ import annotations

from fastapi.testclient import TestClient

from main import app
from store import get_notification, update_notification


def _client() -> TestClient:
    return TestClient(app)


def test_callback_unknown_message_rejected() -> None:
    client = _client()
    response = client.post(
        "/provider-callbacks",
        json={
            "message_id": "missing",
            "provider": "VONAGE",
            "new_state": "Send-success",
            "actual_cost": 0.02,
        },
    )
    assert response.status_code == 200
    data = response.json()["data"]
    assert data["state"] == "rejected"
    assert data["type"] == "unknown_message_id"


def test_callback_valid_transition_accepted() -> None:
    client = _client()
    created = client.post(
        "/notifications",
        json={
            "message_id": "msg-cb-ok",
            "channel_type": "SMS",
            "recipient": "r",
            "content": "c",
            "channel_payload": {"country_code": "US", "phone_number": "+15551230001"},
        },
    ).json()
    nid = created["data"]["notification_id"]
    n = get_notification(nid)
    assert n is not None
    n.state = "Send-to-carrier"
    update_notification(n)

    response = client.post(
        "/provider-callbacks",
        json={
            "message_id": "msg-cb-ok",
            "provider": "VONAGE",
            "new_state": "Send-success",
            "actual_cost": 0.02,
        },
    )
    assert response.status_code == 200
    data = response.json()["data"]
    assert data["state"] == "accepted"
    updated = get_notification(nid)
    assert updated is not None
    assert updated.state == "Send-success"


def test_callback_invalid_transition_rejected() -> None:
    client = _client()
    client.post(
        "/notifications",
        json={
            "message_id": "msg-cb-bad",
            "channel_type": "SMS",
            "recipient": "r",
            "content": "c",
            "channel_payload": {"country_code": "US", "phone_number": "+15551230002"},
        },
    )
    response = client.post(
        "/provider-callbacks",
        json={
            "message_id": "msg-cb-bad",
            "provider": "VONAGE",
            "new_state": "Send-success",
            "actual_cost": 0.01,
        },
    )
    assert response.status_code == 200
    data = response.json()["data"]
    assert data["state"] == "rejected"
    assert data["type"] == "invalid_transition"


def test_callback_duplicate_idempotent() -> None:
    client = _client()
    created = client.post(
        "/notifications",
        json={
            "message_id": "msg-cb-dup",
            "channel_type": "SMS",
            "recipient": "r",
            "content": "c",
            "channel_payload": {"country_code": "US", "phone_number": "+15551230003"},
        },
    ).json()
    nid = created["data"]["notification_id"]
    n = get_notification(nid)
    assert n is not None
    n.state = "Send-to-carrier"
    update_notification(n)

    body = {
        "message_id": "msg-cb-dup",
        "provider": "VONAGE",
        "new_state": "Send-success",
    }
    r1 = client.post("/provider-callbacks", json=body)
    r2 = client.post("/provider-callbacks", json=body)
    assert r1.json()["data"]["state"] == "accepted"
    assert r2.json()["data"]["type"] == "idempotent_no_change"
