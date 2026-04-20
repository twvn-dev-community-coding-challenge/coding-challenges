"""Notification service HTTP API (FastAPI)."""

from __future__ import annotations

import logging
import os
import uuid
from datetime import datetime, timezone

import grpc
from fastapi import Request, status
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse
from google.protobuf import timestamp_pb2
from pydantic import BaseModel, Field

from charging_client import estimate_cost_grpc, record_actual_cost_grpc
from grpc_client import publish_sms_dispatch_via_provider, select_provider
from otp_http_client import OtpIssueError, issue_challenge_http, substitute_otp_in_content
from lifespan import notification_lifespan
from pipeline_runtime import append_pipeline_event as _record_pipeline_event, list_pipeline_events
from models import (
    Notification,
    TransitionEvent,
    TransitionOutcome,
    TransitionSource,
    is_valid_transition,
)
from py_core.app import create_app
from py_core.logging import _should_log_payloads, configure_logging
from store import (
    add_transition_event,
    create_notification,
    find_by_message_id,
    get_notification,
    list_notifications,
    update_notification,
)

logger = logging.getLogger(__name__)


def _pipe(notification_id: str, phase: str, **detail: object) -> None:
    """Append one runtime pipeline row for the tracking UI."""
    payload = {k: v for k, v in detail.items() if v is not None}
    _record_pipeline_event(notification_id, phase, payload)


DEFAULT_MAX_ATTEMPTS = 3

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
    issue_server_otp: bool = Field(
        default=False,
        description="If true, call otp-service to generate a code (hashed at rest with TTL); "
        "substitutes {{OTP}} in content.",
    )


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


async def derive_carrier(phone_number: str) -> str:
    from py_core.db import get_session

    from repository import resolve_carrier

    async with get_session() as session:
        return await resolve_carrier(session, phone_number)


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
        "estimated_cost": notification.estimated_cost,
        "estimated_currency": notification.estimated_currency,
        "charging_estimate_id": notification.charging_estimate_id,
        "charging_rate_id": notification.charging_rate_id,
        "last_actual_cost": notification.last_actual_cost,
        "actual_currency": notification.actual_currency,
        "charging_actual_cost_id": notification.charging_actual_cost_id,
        "otp_challenge_id": notification.otp_challenge_id,
        "otp_expires_at": notification.otp_expires_at,
        "created_at": _datetime_to_iso_z(notification.created_at),
        "updated_at": _datetime_to_iso_z(notification.updated_at),
    }


async def attach_charging_estimate(
    n: Notification,
    *,
    selected_provider_id: str,
    as_of_ts: timestamp_pb2.Timestamp,
) -> JSONResponse | None:
    """Fetch ``EstimateCost`` and persist on *n*. Returns error JSONResponse on gRPC failure."""
    logger.info(
        "notification_charging_estimate_step",
        extra={
            "step": "before_estimate_cost",
            "notification_id": n.notification_id,
            "message_id": n.message_id,
            "provider_id": selected_provider_id,
        },
    )
    try:
        resp = await estimate_cost_grpc(
            message_id=n.message_id,
            provider_id=selected_provider_id,
            country_code=n.channel_payload["country_code"],
            carrier=n.channel_payload["carrier"],
            as_of=as_of_ts,
        )
    except grpc.aio.AioRpcError as err:
        logger.exception(
            "charging_estimate_failed",
            extra={"message_id": n.message_id, "details": err.details()},
        )
        _pipe(
            n.notification_id,
            "grpc.EstimateCost.failed",
            grpc_code=str(err.code()),
            details=err.details(),
        )
        if err.code() == grpc.StatusCode.NOT_FOUND:
            return error_response(
                "CHARGING_RATE_NOT_FOUND",
                "No charging rate for this provider, carrier, and country at as_of",
                status.HTTP_502_BAD_GATEWAY,
                {"grpc": err.details() or err.code().name},
            )
        return error_response(
            "CHARGING_UNAVAILABLE",
            "Charging service failed during cost estimate",
            status.HTTP_502_BAD_GATEWAY,
            {"grpc": err.details() or err.code().name},
        )
    n.estimated_cost = resp.estimated_cost
    n.estimated_currency = resp.currency
    n.charging_estimate_id = resp.estimate_id
    n.charging_rate_id = resp.rate_id
    n.updated_at = datetime.now(timezone.utc)
    update_notification(n)
    logger.info(
        "notification_charging_estimate_step",
        extra={
            "step": "after_estimate_cost_persisted",
            "notification_id": n.notification_id,
            "message_id": n.message_id,
            "charging_estimate_id": n.charging_estimate_id,
            "estimated_cost": n.estimated_cost,
            "currency": n.estimated_currency,
        },
    )
    _pipe(
        n.notification_id,
        "grpc.EstimateCost.ok",
        charging_estimate_id=n.charging_estimate_id,
        estimated_cost=n.estimated_cost,
        currency=n.estimated_currency,
        rate_id=n.charging_rate_id,
    )
    return None


async def record_actual_cost_after_callback(
    n: Notification,
    body: ProviderCallbackRequest,
) -> None:
    """Best-effort ``RecordActualCost`` for terminal billing (does not fail the HTTP callback)."""
    if n.selected_provider_id is None:
        return
    if body.new_state not in ("Send-success", "Send-failed"):
        return
    if body.new_state == "Send-failed" and body.actual_cost is None:
        return

    cost = body.actual_cost
    if cost is None:
        cost = n.estimated_cost
    if cost is None:
        cost = 0.0
    currency = n.estimated_currency or "USD"
    idempotency_key = f"{n.message_id}:actual:{body.new_state}"
    try:
        rec = await record_actual_cost_grpc(
            message_id=n.message_id,
            provider_id=n.selected_provider_id,
            actual_cost=float(cost),
            currency=currency,
            callback_state=body.new_state,
            idempotency_key=idempotency_key,
        )
        n.last_actual_cost = float(cost)
        n.actual_currency = currency
        n.charging_actual_cost_id = rec.actual_cost_id
        n.updated_at = datetime.now(timezone.utc)
        update_notification(n)
    except grpc.aio.AioRpcError:
        logger.exception(
            "charging_record_actual_failed",
            extra={"message_id": n.message_id},
        )


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


_LOGS_DIR = os.path.join(os.path.dirname(__file__), "..", "..", "logs")
configure_logging(service_name="notification-service", log_dir=_LOGS_DIR)

_PAYLOAD_LOGGING = _should_log_payloads(None)

app = create_app(
    title="Notification Service",
    description="SMS lifecycle orchestration; subscribes to sms.dispatch.received (NATS) for Queue → Send-to-carrier",
    service_name="notification-service",
    enable_payload_logging=_PAYLOAD_LOGGING,
    lifespan=notification_lifespan,
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

    content_final = body.content
    otp_challenge_id: str | None = None
    otp_expires_at: str | None = None
    otp_plain_code: str | None = None

    if body.issue_server_otp:
        try:
            otp_payload = await issue_challenge_http(body.message_id)
        except OtpIssueError:
            return error_response(
                "OTP_ISSUE_FAILED",
                "Could not issue server-side OTP (check otp-service and OTP_SERVICE_BASE_URL)",
                status.HTTP_503_SERVICE_UNAVAILABLE,
                {},
            )
        otp_plain_code = str(otp_payload["code"])
        otp_challenge_id = str(otp_payload["challenge_id"])
        otp_expires_at = str(otp_payload["expires_at"])
        content_final = substitute_otp_in_content(body.content, otp_plain_code)

    n = Notification(
        notification_id=nid,
        message_id=body.message_id,
        channel_type=body.channel_type,
        recipient=body.recipient,
        content=content_final,
        channel_payload=payload,
        state="New",
        attempt=0,
        selected_provider_id=None,
        routing_rule_version=None,
        created_at=now,
        updated_at=now,
        otp_challenge_id=otp_challenge_id,
        otp_expires_at=otp_expires_at,
    )
    create_notification(n)
    _pipe(
        nid,
        "http.create_notification",
        message_id=n.message_id,
        state=n.state,
        issue_server_otp=body.issue_server_otp,
    )
    data = notification_to_dict(n)
    expose_plain = os.environ.get("OTP_EXPOSE_PLAINTEXT_TO_CLIENT", "true").lower() in (
        "1",
        "true",
        "yes",
    )
    if body.issue_server_otp and expose_plain and otp_plain_code is not None:
        data["otp_plaintext"] = otp_plain_code
    return success_response(data)


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


@app.get("/notifications/{notification_id}/pipeline-events", response_model=None)
async def get_pipeline_events_endpoint(
    notification_id: str,
) -> dict[str, object] | JSONResponse:
    """Runtime-aggregated SMS pipeline steps for the notification tracking UI."""
    n = get_notification(notification_id)
    if n is None:
        return error_response(
            "NOT_FOUND",
            "Notification not found",
            status.HTTP_404_NOT_FOUND,
            {"notification_id": notification_id},
        )
    return success_response(
        {
            "notification_id": notification_id,
            "message_id": n.message_id,
            "events": list_pipeline_events(notification_id),
        },
    )


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
    carrier = await derive_carrier(n.channel_payload.get("phone_number", ""))
    n.channel_payload = {**n.channel_payload, "carrier": carrier}
    n.updated_at = datetime.now(timezone.utc)
    update_notification(n)

    as_of_ts = iso_string_to_timestamp(body.as_of)
    logger.info(
        "notification_dispatch_flow",
        extra={
            "step": "begin",
            "notification_id": notification_id,
            "message_id": n.message_id,
            "carrier": carrier,
            "as_of": body.as_of,
        },
    )
    _pipe(
        notification_id,
        "dispatch.begin",
        message_id=n.message_id,
        carrier=carrier,
        as_of=body.as_of,
    )
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
        _pipe(
            notification_id,
            "grpc.SelectProvider.failed",
            grpc_code=str(err.code()),
            details=err.details(),
        )
        return error_response(
            "PROVIDER_SELECTION_FAILED",
            "Provider selection failed",
            status.HTTP_502_BAD_GATEWAY,
            {"grpc": err.details() or str(err.code())},
        )

    logger.info(
        "notification_dispatch_flow",
        extra={
            "step": "after_select_provider",
            "notification_id": notification_id,
            "message_id": n.message_id,
            "selected_provider_id": response.selected_provider_id,
            "routing_rule_version": response.routing_rule_version,
        },
    )
    _pipe(
        notification_id,
        "grpc.SelectProvider.ok",
        selected_provider_id=response.selected_provider_id,
        selected_provider_code=response.selected_provider_code,
        routing_rule_version=int(response.routing_rule_version),
    )

    charging_err = await attach_charging_estimate(
        n,
        selected_provider_id=response.selected_provider_id,
        as_of_ts=as_of_ts,
    )
    if charging_err is not None:
        return charging_err

    if n.message_id:
        logger.info(
            "notification_dispatch_flow",
            extra={
                "step": "before_publish_dispatch_requested",
                "notification_id": notification_id,
                "message_id": n.message_id,
            },
        )
        _pipe(notification_id, "dispatch.before_publish_dispatch_requested")
        try:
            pub_resp = await publish_sms_dispatch_via_provider(
                message_id=n.message_id,
                country_code=n.channel_payload["country_code"],
                carrier=carrier,
                selected_provider_id=response.selected_provider_id,
                selected_provider_code=response.selected_provider_code,
                routing_rule_version=int(response.routing_rule_version),
                estimated_cost=float(n.estimated_cost)
                if n.estimated_cost is not None
                else 0.0,
                currency=n.estimated_currency or "",
                charging_estimate_id=n.charging_estimate_id or "",
            )
        except grpc.aio.AioRpcError as err:
            logger.exception(
                "publish_dispatch_requested_failed",
                extra={"notification_id": notification_id, "details": err.details()},
            )
            _pipe(
                notification_id,
                "grpc.PublishSmsDispatchRequested.failed",
                grpc_code=str(err.code()),
                details=err.details(),
            )
            return error_response(
                "DISPATCH_PUBLISH_FAILED",
                "Could not publish sms.dispatch.requested via provider-service",
                status.HTTP_502_BAD_GATEWAY,
                {"grpc": err.details() or str(err.code())},
            )
        if not pub_resp.published:
            _pipe(notification_id, "bus.sms_dispatch.requested.skipped", published=False)
            return error_response(
                "DISPATCH_REQUEST_NOT_PUBLISHED",
                "sms.dispatch.requested was not published; message bus may be unavailable",
                status.HTTP_503_SERVICE_UNAVAILABLE,
                {"message_id": n.message_id},
            )
        logger.info(
            "notification_dispatch_flow",
            extra={
                "step": "after_publish_dispatch_requested",
                "notification_id": notification_id,
                "message_id": n.message_id,
                "published": pub_resp.published,
            },
        )
        _pipe(
            notification_id,
            "grpc.PublishSmsDispatchRequested.ok",
            published=True,
            charging_estimate_id=n.charging_estimate_id,
        )
    else:
        logger.warning(
            "notification_dispatch_skip_bus_publish",
            extra={
                "notification_id": notification_id,
                "reason": "empty_message_id",
            },
        )
        _pipe(notification_id, "dispatch.skip_bus_publish", reason="empty_message_id")

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
    _pipe(notification_id, "state.Send-to-provider", from_state=from_new)

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

    logger.info(
        "notification_dispatch_flow",
        extra={
            "step": "complete_state_queue",
            "notification_id": notification_id,
            "message_id": n.message_id,
            "state": n.state,
            "attempt": n.attempt,
        },
    )
    _pipe(
        notification_id,
        "state.Queue",
        attempt=n.attempt,
        message_id=n.message_id,
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

    if body.as_of is not None:
        as_of_ts = iso_string_to_timestamp(body.as_of)
    else:
        as_of_ts = timestamp_pb2.Timestamp()
        as_of_ts.GetCurrentTime()
    carrier = n.channel_payload.get("carrier")
    if not carrier:
        carrier = await derive_carrier(n.channel_payload.get("phone_number", ""))
        n.channel_payload = {**n.channel_payload, "carrier": carrier}
        n.updated_at = datetime.now(timezone.utc)
        update_notification(n)

    logger.info(
        "notification_retry_flow",
        extra={
            "step": "begin",
            "notification_id": notification_id,
            "message_id": n.message_id,
            "carrier": carrier,
            "current_state": n.state,
            "attempt_before": n.attempt,
        },
    )
    _pipe(
        notification_id,
        "retry.begin",
        carrier=carrier,
        current_state=n.state,
        attempt_before=n.attempt,
    )

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
        _pipe(
            notification_id,
            "grpc.SelectProvider.failed",
            grpc_code=str(err.code()),
            details=err.details(),
        )
        return error_response(
            "PROVIDER_SELECTION_FAILED",
            "Provider selection failed",
            status.HTTP_502_BAD_GATEWAY,
            {"grpc": err.details() or str(err.code())},
        )

    logger.info(
        "notification_retry_flow",
        extra={
            "step": "after_select_provider",
            "notification_id": notification_id,
            "message_id": n.message_id,
            "selected_provider_id": response.selected_provider_id,
            "routing_rule_version": response.routing_rule_version,
        },
    )
    _pipe(
        notification_id,
        "grpc.SelectProvider.ok",
        selected_provider_id=response.selected_provider_id,
        routing_rule_version=int(response.routing_rule_version),
    )

    charging_err = await attach_charging_estimate(
        n,
        selected_provider_id=response.selected_provider_id,
        as_of_ts=as_of_ts,
    )
    if charging_err is not None:
        return charging_err

    if n.message_id:
        logger.info(
            "notification_retry_flow",
            extra={
                "step": "before_publish_dispatch_requested",
                "notification_id": notification_id,
                "message_id": n.message_id,
            },
        )
        _pipe(notification_id, "retry.before_publish_dispatch_requested")
        try:
            pub_resp = await publish_sms_dispatch_via_provider(
                message_id=n.message_id,
                country_code=n.channel_payload["country_code"],
                carrier=carrier,
                selected_provider_id=response.selected_provider_id,
                selected_provider_code=response.selected_provider_code,
                routing_rule_version=int(response.routing_rule_version),
                estimated_cost=float(n.estimated_cost)
                if n.estimated_cost is not None
                else 0.0,
                currency=n.estimated_currency or "",
                charging_estimate_id=n.charging_estimate_id or "",
            )
        except grpc.aio.AioRpcError as err:
            logger.exception(
                "publish_dispatch_requested_failed_retry",
                extra={"notification_id": notification_id, "details": err.details()},
            )
            _pipe(
                notification_id,
                "grpc.PublishSmsDispatchRequested.failed",
                grpc_code=str(err.code()),
                details=err.details(),
            )
            return error_response(
                "DISPATCH_PUBLISH_FAILED",
                "Could not publish sms.dispatch.requested via provider-service",
                status.HTTP_502_BAD_GATEWAY,
                {"grpc": err.details() or str(err.code())},
            )
        if not pub_resp.published:
            _pipe(notification_id, "bus.sms_dispatch.requested.skipped", published=False)
            return error_response(
                "DISPATCH_REQUEST_NOT_PUBLISHED",
                "sms.dispatch.requested was not published; message bus may be unavailable",
                status.HTTP_503_SERVICE_UNAVAILABLE,
                {"message_id": n.message_id},
            )
        logger.info(
            "notification_retry_flow",
            extra={
                "step": "after_publish_dispatch_requested",
                "notification_id": notification_id,
                "message_id": n.message_id,
                "published": pub_resp.published,
            },
        )
        _pipe(
            notification_id,
            "grpc.PublishSmsDispatchRequested.ok",
            published=True,
            charging_estimate_id=n.charging_estimate_id,
        )
    else:
        logger.warning(
            "notification_retry_skip_bus_publish",
            extra={
                "notification_id": notification_id,
                "reason": "empty_message_id",
            },
        )
        _pipe(notification_id, "retry.skip_bus_publish", reason="empty_message_id")

    from_state = n.state
    n.state = "Send-to-provider"
    n.attempt += 1
    n.selected_provider_id = response.selected_provider_id
    n.routing_rule_version = int(response.routing_rule_version)
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
    _pipe(notification_id, "state.Send-to-provider", from_state=from_state)

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
    _pipe(
        notification_id,
        "state.Queue",
        attempt=n.attempt,
        message_id=n.message_id,
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

    await record_actual_cost_after_callback(n, body)

    return success_response(
        {
            "state": "accepted",
            "type": "applied",
            "reason": "callback_applied",
        }
    )
