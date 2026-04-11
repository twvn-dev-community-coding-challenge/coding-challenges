"""Notification service HTTP API (FastAPI)."""

from __future__ import annotations

import logging
import uuid
from datetime import datetime, timezone

import grpc
from fastapi import Request, status
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse
from google.protobuf import timestamp_pb2
from pydantic import BaseModel, Field

from grpc_client import select_provider
from models import (
    Notification,
    TransitionEvent,
    TransitionOutcome,
    TransitionSource,
    is_valid_transition,
)
from py_core.app import create_app
from store import (
    add_transition_event,
    create_notification,
    find_by_message_id,
    get_notification,
    list_notifications,
    update_notification,
)

logger = logging.getLogger(__name__)

DEFAULT_MAX_ATTEMPTS = 3

PHONE_PREFIX_TO_CARRIER: dict[str, str] = {
    "+84": "VIETTEL",
    "+1": "T-MOBILE",
    "+44": "VODAFONE",
    "+61": "TELSTRA",
}

RETRYABLE_STATES: frozenset[str] = frozenset({"Send-failed", "Carrier-rejected"})


class SmsChannelPayload(BaseModel):
    country_code: str
    phone_number: str


class CreateNotificationRequest(BaseModel):
    message_id: str
    channel_type: str = "SMS"
    recipient: str
    content: str
    channel_payload: SmsChannelPayload


class DispatchRequest(BaseModel):
    as_of: str = Field(..., description="ISO 8601 datetime string")


class RetryRequest(BaseModel):
    as_of: str | None = None


class ProviderCallbackRequest(BaseModel):
    message_id: str
    provider: str
    new_state: str
    actual_cost: float | None = None


def success_response(data: dict[str, object]) -> dict[str, object]:
    return {"data": data}


def error_response(
    code: str,
    message: str,
    status_code: int,
    details: dict[str, object] | None = None,
) -> JSONResponse:
    payload: dict[str, object] = {
        "error": {
            "code": code,
            "message": message,
            "details": details if details is not None else {},
        }
    }
    return JSONResponse(status_code=status_code, content=payload)


def derive_carrier(phone_number: str) -> str:
    for prefix, carrier in PHONE_PREFIX_TO_CARRIER.items():
        if phone_number.startswith(prefix):
            return carrier
    return "UNKNOWN"


def _datetime_to_iso_z(dt: datetime) -> str:
    if dt.tzinfo is None:
        dt = dt.replace(tzinfo=timezone.utc)
    return dt.astimezone(timezone.utc).isoformat().replace("+00:00", "Z")


def iso_string_to_timestamp(iso: str) -> timestamp_pb2.Timestamp:
    normalized = iso.replace("Z", "+00:00")
    parsed = datetime.fromisoformat(normalized)
    if parsed.tzinfo is None:
        parsed = parsed.replace(tzinfo=timezone.utc)
    ts = timestamp_pb2.Timestamp()
    ts.FromDatetime(parsed.astimezone(timezone.utc))
    return ts


def notification_to_dict(notification: Notification) -> dict[str, object]:
    return {
        "notification_id": notification.notification_id,
        "message_id": notification.message_id,
        "channel_type": notification.channel_type,
        "recipient": notification.recipient,
        "content": notification.content,
        "channel_payload": dict(notification.channel_payload),
        "state": notification.state,
        "attempt": notification.attempt,
        "selected_provider_id": notification.selected_provider_id,
        "routing_rule_version": notification.routing_rule_version,
        "created_at": _datetime_to_iso_z(notification.created_at),
        "updated_at": _datetime_to_iso_z(notification.updated_at),
    }


def _record_transition(
    notification_id: str,
    from_state: str,
    to_state: str,
    source: TransitionSource,
    outcome: TransitionOutcome,
    reason: str,
) -> None:
    now = datetime.now(timezone.utc)
    add_transition_event(
        TransitionEvent(
            notification_id=notification_id,
            from_state=from_state,
            to_state=to_state,
            at=now,
            source=source,
            outcome=outcome,
            reason=reason,
        )
    )


app = create_app(
    title="Notification Service",
    description="SMS lifecycle orchestration",
)


@app.exception_handler(RequestValidationError)
async def validation_exception_handler(
    request: Request,
    exc: RequestValidationError,
) -> JSONResponse:
    _ = request
    return error_response(
        "VALIDATION_ERROR",
        "Request validation failed",
        status.HTTP_422_UNPROCESSABLE_CONTENT,
        {"errors": exc.errors()},
    )


@app.post(
    "/notifications",
    status_code=status.HTTP_201_CREATED,
    response_model=None,
)
async def create_notification_endpoint(
    body: CreateNotificationRequest,
) -> dict[str, object] | JSONResponse:
    if body.channel_type != "SMS":
        return error_response(
            "VALIDATION_ERROR",
            "Only SMS channel is supported in phase 1",
            status.HTTP_422_UNPROCESSABLE_CONTENT,
        )
    if find_by_message_id(body.message_id) is not None:
        return error_response(
            "VALIDATION_ERROR",
            "message_id already exists",
            status.HTTP_422_UNPROCESSABLE_CONTENT,
            {"message_id": body.message_id},
        )
    now = datetime.now(timezone.utc)
    nid = str(uuid.uuid4())
    payload = {
        "country_code": body.channel_payload.country_code,
        "phone_number": body.channel_payload.phone_number,
    }
    n = Notification(
        notification_id=nid,
        message_id=body.message_id,
        channel_type=body.channel_type,
        recipient=body.recipient,
        content=body.content,
        channel_payload=payload,
        state="New",
        attempt=0,
        selected_provider_id=None,
        routing_rule_version=None,
        created_at=now,
        updated_at=now,
    )
    create_notification(n)
    return success_response(notification_to_dict(n))


@app.get("/notifications/{notification_id}", response_model=None)
async def get_notification_endpoint(
    notification_id: str,
) -> dict[str, object] | JSONResponse:
    n = get_notification(notification_id)
    if n is None:
        return error_response(
            "NOT_FOUND",
            "Notification not found",
            status.HTTP_404_NOT_FOUND,
            {"notification_id": notification_id},
        )
    return success_response(notification_to_dict(n))


@app.get("/notifications")
async def list_notifications_endpoint() -> dict[str, object]:
    items = sorted(
        list_notifications(),
        key=lambda x: x.created_at,
        reverse=True,
    )
    return success_response({"notifications": [notification_to_dict(n) for n in items]})


@app.post("/notifications/{notification_id}/dispatch", response_model=None)
async def dispatch_notification(
    notification_id: str,
    body: DispatchRequest,
) -> dict[str, object] | JSONResponse:
    n = get_notification(notification_id)
    if n is None:
        return error_response(
            "NOT_FOUND",
            "Notification not found",
            status.HTTP_404_NOT_FOUND,
            {"notification_id": notification_id},
        )
    if n.state != "New":
        return error_response(
            "INVALID_STATE_TRANSITION",
            "Dispatch is only allowed from New state",
            status.HTTP_409_CONFLICT,
            {"current_state": n.state},
        )
    carrier = derive_carrier(n.channel_payload.get("phone_number", ""))
    n.channel_payload = {**n.channel_payload, "carrier": carrier}
    n.updated_at = datetime.now(timezone.utc)
    update_notification(n)

    as_of_ts = iso_string_to_timestamp(body.as_of)
    try:
        response = await select_provider(
            country_code=n.channel_payload["country_code"],
            carrier=carrier,
            as_of=as_of_ts,
            policy="highest_precedence",
            message_id=n.message_id,
        )
    except grpc.aio.AioRpcError as err:
        logger.exception(
            "provider_selection_failed",
            extra={"notification_id": notification_id, "details": err.details()},
        )
        return error_response(
            "PROVIDER_SELECTION_FAILED",
            "Provider selection failed",
            status.HTTP_502_BAD_GATEWAY,
            {"grpc": err.details() or str(err.code())},
        )

    from_new = n.state
    n.state = "Send-to-provider"
    n.selected_provider_id = response.selected_provider_id
    n.routing_rule_version = int(response.routing_rule_version)
    n.updated_at = datetime.now(timezone.utc)
    update_notification(n)
    _record_transition(
        n.notification_id,
        from_new,
        "Send-to-provider",
        "system",
        "accepted",
        "dispatch_new_to_send_to_provider",
    )

    n.attempt = 1
    from_sp = n.state
    n.state = "Queue"
    n.updated_at = datetime.now(timezone.utc)
    update_notification(n)
    _record_transition(
        n.notification_id,
        from_sp,
        "Queue",
        "system",
        "accepted",
        "dispatch_send_to_provider_to_queue",
    )

    return success_response(notification_to_dict(n))


@app.post("/notifications/{notification_id}/retry", response_model=None)
async def retry_notification(
    notification_id: str,
    body: RetryRequest,
) -> dict[str, object] | JSONResponse:
    n = get_notification(notification_id)
    if n is None:
        return error_response(
            "NOT_FOUND",
            "Notification not found",
            status.HTTP_404_NOT_FOUND,
            {"notification_id": notification_id},
        )
    if n.state not in RETRYABLE_STATES:
        return error_response(
            "RETRY_NOT_ALLOWED",
            "Notification is not in a retryable state",
            status.HTTP_409_CONFLICT,
            {"current_state": n.state},
        )
    if n.attempt >= DEFAULT_MAX_ATTEMPTS:
        return error_response(
            "RETRY_NOT_ALLOWED",
            "Maximum retry attempts reached",
            status.HTTP_409_CONFLICT,
            {"attempt": n.attempt, "max_attempts": DEFAULT_MAX_ATTEMPTS},
        )

    from_state = n.state
    n.state = "Send-to-provider"
    n.attempt += 1
    n.updated_at = datetime.now(timezone.utc)
    update_notification(n)
    _record_transition(
        n.notification_id,
        from_state,
        "Send-to-provider",
        "retry",
        "accepted",
        "retry_to_send_to_provider",
    )

    if body.as_of is not None:
        as_of_ts = iso_string_to_timestamp(body.as_of)
    else:
        as_of_ts = timestamp_pb2.Timestamp()
        as_of_ts.GetCurrentTime()
    carrier = n.channel_payload.get("carrier")
    if not carrier:
        carrier = derive_carrier(n.channel_payload.get("phone_number", ""))
        n.channel_payload = {**n.channel_payload, "carrier": carrier}
    try:
        response = await select_provider(
            country_code=n.channel_payload["country_code"],
            carrier=carrier,
            as_of=as_of_ts,
            policy="highest_precedence",
            message_id=n.message_id,
        )
    except grpc.aio.AioRpcError as err:
        logger.exception(
            "provider_selection_failed_retry",
            extra={"notification_id": notification_id, "details": err.details()},
        )
        return error_response(
            "PROVIDER_SELECTION_FAILED",
            "Provider selection failed",
            status.HTTP_502_BAD_GATEWAY,
            {"grpc": err.details() or str(err.code())},
        )

    n.selected_provider_id = response.selected_provider_id
    n.routing_rule_version = int(response.routing_rule_version)
    n.updated_at = datetime.now(timezone.utc)
    update_notification(n)

    from_sp = n.state
    n.state = "Queue"
    n.updated_at = datetime.now(timezone.utc)
    update_notification(n)
    _record_transition(
        n.notification_id,
        from_sp,
        "Queue",
        "system",
        "accepted",
        "retry_send_to_provider_to_queue",
    )

    return success_response(notification_to_dict(n))


@app.post("/provider-callbacks")
async def provider_callbacks(
    body: ProviderCallbackRequest,
) -> dict[str, object]:
    n = find_by_message_id(body.message_id)
    if n is None:
        return success_response(
            {
                "state": "rejected",
                "type": "unknown_message_id",
                "reason": "messageId_not_found",
            }
        )

    if n.state == body.new_state:
        return success_response(
            {
                "state": "accepted",
                "type": "idempotent_no_change",
                "reason": "duplicate_callback_already_applied",
            }
        )

    if not is_valid_transition(n.state, body.new_state):
        return success_response(
            {
                "state": "rejected",
                "type": "invalid_transition",
                "reason": "transition_not_allowed_from_current_state",
            }
        )

    from_state = n.state
    n.state = body.new_state
    n.updated_at = datetime.now(timezone.utc)
    update_notification(n)
    _record_transition(
        n.notification_id,
        from_state,
        body.new_state,
        "callback",
        "accepted",
        f"callback_from_{body.provider}",
    )

    return success_response(
        {
            "state": "accepted",
            "type": "applied",
            "reason": "callback_applied",
        }
    )
