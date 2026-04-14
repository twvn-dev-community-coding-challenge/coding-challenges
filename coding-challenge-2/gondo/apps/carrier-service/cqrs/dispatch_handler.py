"""CQRS: consume ``sms.dispatch.requested``, call provider HTTP API, publish outcome."""

from __future__ import annotations

import json
import logging
from dataclasses import dataclass

import httpx

from bus_state import get_message_bus
from cqrs.rate_gate import acquire_send_slot, release_send_slot
from py_core.bus.contract import publish_json
from py_core.bus.topics import SMS_DISPATCH_OUTCOME, SMS_DISPATCH_RECEIVED

logger = logging.getLogger(__name__)


@dataclass(frozen=True)
class _DispatchPayload:
    message_id: str
    correlation_id: str
    country_code: str
    carrier: str
    provider_id: str
    provider_code: str
    api_endpoint: str


async def handle_dispatch_requested(raw: bytes) -> None:
    try:
        data = json.loads(raw.decode("utf-8"))
    except (json.JSONDecodeError, UnicodeDecodeError):
        logger.warning("dispatch_malformed", extra={"raw": payload_preview(raw)})
        return
    bus = get_message_bus()
    if bus is None:
        return
    payload = _DispatchPayload(
        message_id=str(data.get("message_id", "")),
        correlation_id=str(data.get("correlation_id", "")),
        country_code=str(data.get("country_code", "")),
        carrier=str(data.get("carrier", "")),
        provider_id=str(data.get("provider_id", "")),
        provider_code=str(data.get("provider_code", "")),
        api_endpoint=str(data.get("api_endpoint", "")),
    )
    await acquire_send_slot()
    try:
        await publish_json(
            bus,
            SMS_DISPATCH_RECEIVED,
            {
                "message_id": payload.message_id,
                "correlation_id": payload.correlation_id,
                "country_code": payload.country_code,
                "carrier": payload.carrier,
                "provider_id": payload.provider_id,
                "provider_code": payload.provider_code,
            },
        )
        logger.info(
            "sms_dispatch_received_published",
            extra={"message_id": payload.message_id},
        )
        try:
            ok, detail = await _try_http_probe(payload.api_endpoint)
            status = "success" if ok else "failure"
        except Exception as exc:
            status = "failure"
            detail = str(exc)[:500]
        outcome = {
            "status": status,
            "message_id": payload.message_id,
            "correlation_id": payload.correlation_id,
            "detail": detail,
            "provider_id": payload.provider_id,
        }
        await publish_json(bus, SMS_DISPATCH_OUTCOME, outcome)
        logger.info(
            "sms_dispatch_outcome_published",
            extra={"status": outcome["status"], "message_id": payload.message_id},
        )
    finally:
        release_send_slot()


def payload_preview(raw: bytes, n: int = 200) -> str:
    try:
        return raw[:n].decode("utf-8", errors="replace")
    except Exception:
        return repr(raw[:n])


async def _try_http_probe(api_endpoint: str) -> tuple[bool, str]:
    """Skeleton: GET to registry base URL — replace with real SMS API client."""
    if not api_endpoint or not api_endpoint.startswith("http"):
        return False, "invalid_api_endpoint"
    try:
        async with httpx.AsyncClient(timeout=5.0, follow_redirects=True) as client:
            r = await client.get(api_endpoint)
            if r.status_code < 500:
                return True, f"http_status_{r.status_code}"
    except httpx.HTTPError as exc:
        return False, str(exc)[:500]
    return False, "http_failed"
