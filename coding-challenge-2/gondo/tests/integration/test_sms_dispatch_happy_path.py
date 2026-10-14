"""Happy-path SMS flow against running infra: notification → provider gRPC → NATS → carrier → notification state.

Requires **PostgreSQL** (routing + carrier prefix), **NATS**, **provider-service**, **carrier-service**,
**notification-service** — e.g. `docker compose up -d` from `gondo/`.

Run (use **-s** so step logs print)::

    cd gondo
    RUN_INTEGRATION_TESTS=1 python -m pytest tests/integration -v -s

Optional: ``NOTIFICATION_SERVICE_URL`` (default ``http://127.0.0.1:8001``).
"""

from __future__ import annotations

import logging
import time
import uuid

import httpx
import pytest

logger = logging.getLogger("integration")

# VN + Viettel: seed maps 84+9[6-8] to VIETTEL; 84+90 is MOBIFONE (see provider 0002_seed_data).
_PHONE = "+84961234567"
_COUNTRY = "VN"
_AS_OF = "2026-06-15T12:00:00.000Z"
_POLL_INTERVAL_S = 0.5
_MAX_WAIT_SEND_TO_CARRIER_S = 25.0


def _log(msg: str) -> None:
    line = f"[integration] {msg}"
    print(line, flush=True)
    logger.info("%s", msg)


def _get_notification(client: httpx.Client, base: str, nid: str) -> dict:
    r = client.get(f"{base}/notifications/{nid}", timeout=10.0)
    r.raise_for_status()
    return r.json()["data"]


def _wait_for_state(
    client: httpx.Client,
    base: str,
    nid: str,
    want: str,
    timeout_s: float,
) -> dict:
    deadline = time.monotonic() + timeout_s
    started = time.monotonic()
    last: dict | None = None
    poll_n = 0
    _log(f"Polling GET /notifications/{nid} until state={want!r} (timeout {timeout_s}s, every {_POLL_INTERVAL_S}s)")
    while time.monotonic() < deadline:
        last = _get_notification(client, base, nid)
        poll_n += 1
        st = last.get("state")
        if st == want:
            elapsed = time.monotonic() - started
            _log(f"Reached state {want!r} after {poll_n} poll(s) in {elapsed:.1f}s")
            return last
        if poll_n == 1 or poll_n % 10 == 0:
            _log(f"  poll #{poll_n}: state={st!r}")
        time.sleep(_POLL_INTERVAL_S)
    assert last is not None
    pytest.fail(
        f"Timed out waiting for state {want!r}; last state={last.get('state')!r}",
    )


@pytest.mark.integration
def test_sms_dispatch_travels_to_carrier_and_back_to_notification_queue_then_send_to_carrier(
    integration_session: str,
) -> None:
    """Create SMS → dispatch → Queue; async path sets Send-to-carrier after carrier publishes received."""
    base = integration_session
    message_id = f"itest-{uuid.uuid4().hex[:12]}"
    _log("=== SMS integration happy path ===")
    _log(f"message_id={message_id} phone={_PHONE} country={_COUNTRY} as_of={_AS_OF}")

    with httpx.Client(timeout=30.0) as client:
        _log("Step 1: POST /notifications (create SMS notification in New)")
        create = client.post(
            f"{base}/notifications",
            json={
                "message_id": message_id,
                "channel_type": "SMS",
                "recipient": "integration-user",
                "content": "integration hello",
                "channel_payload": {
                    "country_code": _COUNTRY,
                    "phone_number": _PHONE,
                },
            },
        )
        _log(f"  → HTTP {create.status_code}")
        assert create.status_code == 201, create.text
        nid = create.json()["data"]["notification_id"]
        assert create.json()["data"]["state"] == "New"
        _log(f"  → notification_id={nid} state=New")

        _log(
            "Step 2: POST /notifications/{id}/dispatch "
            "(SelectProvider → EstimateCost → PublishSmsDispatchRequested → NATS sms.dispatch.requested; "
            "expect state Queue)",
        )
        dispatch = client.post(
            f"{base}/notifications/{nid}/dispatch",
            json={"as_of": _AS_OF},
        )
        _log(f"  → HTTP {dispatch.status_code}")
        assert dispatch.status_code == 200, dispatch.text
        body = dispatch.json()["data"]
        assert body["state"] == "Queue"
        assert body["selected_provider_id"]
        assert body["channel_payload"].get("carrier") == "VIETTEL"
        _log(
            f"  → state=Queue selected_provider_id={body['selected_provider_id']} "
            f"routing_rule_version={body.get('routing_rule_version')} carrier={body['channel_payload'].get('carrier')}",
        )

        _log(
            "Step 3: Wait for async path (carrier consumes requested → publishes received → "
            "notification subscriber → Queue → Send-to-carrier)",
        )
        got = _wait_for_state(
            client,
            base,
            nid,
            "Send-to-carrier",
            _MAX_WAIT_SEND_TO_CARRIER_S,
        )
        assert got["message_id"] == message_id
        _log(f"=== Done: final state={got.get('state')!r} message_id={got.get('message_id')} ===")
