"""Queue → Send-to-carrier → Carrier-rejected (MNO sim) → retry — User Story 4 trace."""

from __future__ import annotations

from datetime import datetime, timezone
from unittest.mock import AsyncMock, patch

from fastapi.testclient import TestClient

from generated import charging_pb2, provider_pb2
from py_core.proto_utils import datetime_to_timestamp

from cqrs.carrier_dispatch_received import apply_carrier_dispatch_received
from main import app
from pipeline_runtime import list_pipeline_events
from models import Notification
from store import clear, create_notification, get_notification, get_transition_events


def _estimate() -> charging_pb2.EstimateCostResponse:
    er = charging_pb2.EstimateCostResponse()
    er.estimate_id = "est-cr"
    er.estimated_cost = 0.02
    er.currency = "USD"
    er.rate_id = "rate_cr"
    er.rate_version = 1
    er.created_at.CopyFrom(datetime_to_timestamp(datetime.now(timezone.utc)))
    return er


def test_queue_to_carrier_rejected_then_retry_round_trip() -> None:
    """0395000005 (mock_scenarios carrier_rejected): bus ack → Send-to-carrier → Carrier-rejected; retry → Queue."""
    clear()
    client = TestClient(app)

    created = client.post(
        "/notifications",
        json={
            "message_id": "msg-us4-cr",
            "channel_type": "SMS",
            "recipient": "u",
            "content": "hello",
            "channel_payload": {"country_code": "VN", "phone_number": "0395000005"},
        },
    ).json()
    nid = created["data"]["notification_id"]

    mock_resp = provider_pb2.SelectProviderResponse()
    mock_resp.selected_provider_id = "prv_01"
    mock_resp.selected_provider_code = "TWILIO"
    mock_resp.routing_rule_version = 2
    mock_pub = provider_pb2.PublishSmsDispatchRequestedResponse()
    mock_pub.published = True

    with (
        patch("cqrs.dispatch_pipeline.select_provider", new_callable=AsyncMock, return_value=mock_resp),
        patch(
            "cqrs.dispatch_pipeline.publish_sms_dispatch_via_provider",
            new_callable=AsyncMock,
            return_value=mock_pub,
        ),
        patch(
            "main.derive_carrier",
            new_callable=AsyncMock,
            return_value="MOBIFONE",
        ),
        patch(
            "cqrs.dispatch_pipeline.estimate_cost_grpc",
            new_callable=AsyncMock,
            return_value=_estimate(),
        ),
    ):
        disp = client.post(
            f"/notifications/{nid}/dispatch",
            json={"as_of": "2026-04-03T12:00:00.000Z"},
        )

    assert disp.status_code == 200
    assert disp.json()["data"]["state"] == "Queue"

    assert apply_carrier_dispatch_received("msg-us4-cr") == "applied_carrier_rejected"
    after_reject = get_notification(nid)
    assert after_reject is not None
    assert after_reject.state == "Carrier-rejected"

    transitions = [(e.from_state, e.to_state) for e in get_transition_events(nid)]
    assert ("Queue", "Send-to-carrier") in transitions
    assert ("Send-to-carrier", "Carrier-rejected") in transitions

    phases = [e["phase"] for e in list_pipeline_events(nid)]
    assert "state.Send-to-carrier" in phases
    assert "state.Carrier-rejected" in phases

    with (
        patch("cqrs.dispatch_pipeline.select_provider", new_callable=AsyncMock, return_value=mock_resp),
        patch(
            "cqrs.dispatch_pipeline.publish_sms_dispatch_via_provider",
            new_callable=AsyncMock,
            return_value=mock_pub,
        ),
        patch(
            "cqrs.dispatch_pipeline.estimate_cost_grpc",
            new_callable=AsyncMock,
            return_value=_estimate(),
        ),
    ):
        retry = client.post(
            f"/notifications/{nid}/retry",
            json={"as_of": "2026-04-03T12:00:00.000Z"},
        )

    assert retry.status_code == 200
    assert retry.json()["data"]["state"] == "Queue"

    transitions_after = [(e.from_state, e.to_state) for e in get_transition_events(nid)]
    assert ("Carrier-rejected", "Send-to-provider") in transitions_after
    assert ("Send-to-provider", "Queue") in transitions_after

    assert apply_carrier_dispatch_received("msg-us4-cr") == "applied_carrier_rejected"
    assert get_notification(nid).state == "Carrier-rejected"


def test_carrier_dispatch_received_magic_number_mno_reject_after_send_to_carrier() -> None:
    """VN MSISDN in mock_scenarios (carrier_rejected): Queue → Send-to-carrier → Carrier-rejected."""
    clear()
    now = datetime.now(timezone.utc)
    n = Notification(
        notification_id="nid-magic",
        message_id="msg-magic-skip",
        channel_type="SMS",
        recipient="u",
        content="hello",
        channel_payload={"country_code": "VN", "phone_number": "0395000005"},
        state="Queue",
        attempt=1,
        selected_provider_id="prv_01",
        selected_provider_code=None,
        routing_rule_version=2,
        created_at=now,
        updated_at=now,
    )
    create_notification(n)

    assert apply_carrier_dispatch_received("msg-magic-skip") == "applied_carrier_rejected"
    updated = get_notification("nid-magic")
    assert updated is not None
    assert updated.state == "Carrier-rejected"
