"""Dispatch endpoint tests."""

from __future__ import annotations

import grpc
from datetime import datetime, timezone
from unittest.mock import AsyncMock, patch

from fastapi.testclient import TestClient

from generated import charging_pb2, provider_pb2
from py_core.proto_utils import datetime_to_timestamp
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


def _sample_estimate_response() -> charging_pb2.EstimateCostResponse:
    er = charging_pb2.EstimateCostResponse()
    er.estimate_id = "est-test-1"
    er.estimated_cost = 0.015
    er.currency = "USD"
    er.rate_id = "rate_test"
    er.rate_version = 1
    er.created_at.CopyFrom(datetime_to_timestamp(datetime.now(timezone.utc)))
    return er


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
    mock_resp.selected_provider_id = "prv_01"
    mock_resp.routing_rule_version = 2
    mock_resp.sms_dispatch_requested_published = True

    with (
        patch("main.select_provider", new_callable=AsyncMock, return_value=mock_resp),
        patch(
            "repository.resolve_carrier", new_callable=AsyncMock, return_value="VIETTEL"
        ),
        patch(
            "main.estimate_cost_grpc",
            new_callable=AsyncMock,
            return_value=_sample_estimate_response(),
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
    assert data["selected_provider_id"] == "prv_01"
    assert data["routing_rule_version"] == 2
    assert data["channel_payload"]["carrier"] == "VIETTEL"
    assert data["estimated_cost"] == 0.015
    assert data["charging_estimate_id"] == "est-test-1"


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


def test_dispatch_rejects_when_notification_not_in_new_state() -> None:
    """Dispatch is only allowed from New; second call must return 409."""
    client = TestClient(app)
    created = client.post(
        "/notifications",
        json={
            "message_id": "msg-double-dispatch",
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
            "main.derive_carrier",
            new_callable=AsyncMock,
            return_value="VIETTEL",
        ),
        patch(
            "main.estimate_cost_grpc",
            new_callable=AsyncMock,
            return_value=_sample_estimate_response(),
        ),
    ):
        first = client.post(
            f"/notifications/{nid}/dispatch",
            json={"as_of": "2026-04-03T12:00:00.000Z"},
        )
        assert first.status_code == 200

        second = client.post(
            f"/notifications/{nid}/dispatch",
            json={"as_of": "2026-04-03T12:00:00.000Z"},
        )

    assert second.status_code == 409
    assert second.json()["error"]["code"] == "INVALID_STATE_TRANSITION"


async def _raise_grpc_not_found(*_a, **_k):
    raise grpc.aio.AioRpcError(grpc.StatusCode.NOT_FOUND, None, None)


async def _raise_grpc_unavailable(*_a, **_k):
    raise grpc.aio.AioRpcError(grpc.StatusCode.UNAVAILABLE, None, None)


def test_dispatch_returns_502_when_charging_rate_not_found() -> None:
    client = TestClient(app)
    created = client.post(
        "/notifications",
        json={
            "message_id": "msg-charge-nf",
            "channel_type": "SMS",
            "recipient": "u",
            "content": "hello",
            "channel_payload": {"country_code": "US", "phone_number": "+84901234567"},
        },
    ).json()
    nid = created["data"]["notification_id"]

    mock_resp = provider_pb2.SelectProviderResponse()
    mock_resp.selected_provider_id = "prv_01"
    mock_resp.routing_rule_version = 2
    mock_resp.sms_dispatch_requested_published = True

    with (
        patch("main.select_provider", new_callable=AsyncMock, return_value=mock_resp),
        patch(
            "main.derive_carrier",
            new_callable=AsyncMock,
            return_value="VIETTEL",
        ),
        patch("main.estimate_cost_grpc", _raise_grpc_not_found),
    ):
        response = client.post(
            f"/notifications/{nid}/dispatch",
            json={"as_of": "2026-04-03T12:00:00.000Z"},
        )

    assert response.status_code == 502
    assert response.json()["error"]["code"] == "CHARGING_RATE_NOT_FOUND"


def test_dispatch_returns_502_when_charging_service_unavailable() -> None:
    client = TestClient(app)
    created = client.post(
        "/notifications",
        json={
            "message_id": "msg-charge-unavail",
            "channel_type": "SMS",
            "recipient": "u",
            "content": "hello",
            "channel_payload": {"country_code": "US", "phone_number": "+84901234567"},
        },
    ).json()
    nid = created["data"]["notification_id"]

    mock_resp = provider_pb2.SelectProviderResponse()
    mock_resp.selected_provider_id = "prv_01"
    mock_resp.routing_rule_version = 2
    mock_resp.sms_dispatch_requested_published = True

    with (
        patch("main.select_provider", new_callable=AsyncMock, return_value=mock_resp),
        patch(
            "main.derive_carrier",
            new_callable=AsyncMock,
            return_value="VIETTEL",
        ),
        patch("main.estimate_cost_grpc", _raise_grpc_unavailable),
    ):
        response = client.post(
            f"/notifications/{nid}/dispatch",
            json={"as_of": "2026-04-03T12:00:00.000Z"},
        )

    assert response.status_code == 502
    assert response.json()["error"]["code"] == "CHARGING_UNAVAILABLE"


def test_dispatch_returns_404_for_unknown_notification() -> None:
    client = TestClient(app)
    response = client.post(
        "/notifications/00000000-0000-0000-0000-000000000099/dispatch",
        json={"as_of": "2026-04-03T12:00:00.000Z"},
    )
    assert response.status_code == 404
    assert response.json()["error"]["code"] == "NOT_FOUND"


def test_retry_returns_404_for_unknown_notification() -> None:
    client = TestClient(app)
    response = client.post(
        "/notifications/00000000-0000-0000-0000-000000000088/retry",
        json={},
    )
    assert response.status_code == 404
    assert response.json()["error"]["code"] == "NOT_FOUND"


def test_retry_returns_409_when_notification_not_retryable() -> None:
    """Retry is only allowed from Send-failed / Carrier-rejected."""
    client = TestClient(app)
    created = client.post(
        "/notifications",
        json={
            "message_id": "msg-retry-new-only",
            "channel_type": "SMS",
            "recipient": "u",
            "content": "hello",
            "channel_payload": {"country_code": "US", "phone_number": "+84901234567"},
        },
    ).json()
    nid = created["data"]["notification_id"]

    response = client.post(f"/notifications/{nid}/retry", json={})

    assert response.status_code == 409
    err = response.json()["error"]
    assert err["code"] == "RETRY_NOT_ALLOWED"
    assert err["details"]["current_state"] == "New"


def test_dispatch_returns_502_when_provider_grpc_unavailable() -> None:
    """Provider SelectProvider fails (simulated infra / gRPC UNAVAILABLE)."""
    client = TestClient(app)
    created = client.post(
        "/notifications",
        json={
            "message_id": "msg-prov-down",
            "channel_type": "SMS",
            "recipient": "u",
            "content": "hello",
            "channel_payload": {"country_code": "US", "phone_number": "+84901234567"},
        },
    ).json()
    nid = created["data"]["notification_id"]

    with (
        patch(
            "main.select_provider",
            AsyncMock(
                side_effect=grpc.aio.AioRpcError(
                    grpc.StatusCode.UNAVAILABLE,
                    None,
                    None,
                ),
            ),
        ),
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

    assert response.status_code == 502
    assert response.json()["error"]["code"] == "PROVIDER_SELECTION_FAILED"
