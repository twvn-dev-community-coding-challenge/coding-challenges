"""Notification service HTTP API (FastAPI)."""

from __future__ import annotations

import logging
import os
import uuid
from datetime import datetime, timezone
from typing import Annotated

from fastapi import Header, Query, Request, status
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse
from google.protobuf import timestamp_pb2
from cqrs.charging_callbacks import record_actual_cost_after_callback
from cqrs.dispatch_pipeline import iso_string_to_timestamp, orchestrate_select_charge_publish
from cqrs.transitions import record_transition
from kpis import build_sms_kpis, parse_iso_datetime
from lifespan import notification_lifespan
from otp_http_client import OtpIssueError, issue_challenge_http, substitute_otp_in_content
from pipeline_runtime import list_pipeline_events, pipe
from py_core.app import create_app
from py_core.logging import _should_log_payloads, configure_logging
from responses import error_response, success_response
from schemas import (
    CreateNotificationRequest,
    DispatchRequest,
    ProviderCallbackRequest,
    RetryRequest,
)
from serialization import notification_to_dict
from store import (
    create_notification,
    find_by_message_id,
    get_notification,
    list_notifications,
    update_notification,
)
from models import Notification, is_valid_transition

logger = logging.getLogger(__name__)

DEFAULT_MAX_ATTEMPTS = 3

_CALLING_DOMAIN_MAX_LEN = 128


def _normalize_x_calling_domain(raw: str | None) -> str | None:
    """Return stripped caller label for US2-style analytics, or None if absent/blank."""
    if raw is None:
        return None
    s = raw.strip()
    if not s:
        return None
    if len(s) > _CALLING_DOMAIN_MAX_LEN:
        raise ValueError(
            f"X-Calling-Domain must be at most {_CALLING_DOMAIN_MAX_LEN} characters"
        )
    if any(ch in s for ch in ("\n", "\r", "\x00")):
        raise ValueError("X-Calling-Domain contains invalid characters")
    return s

RETRYABLE_STATES: frozenset[str] = frozenset({"Send-failed", "Carrier-rejected"})


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


async def derive_carrier(phone_number: str) -> str:
    from py_core.db import get_session

    from repository import resolve_carrier

    async with get_session() as session:
        return await resolve_carrier(session, phone_number)


@app.post(
    "/notifications",
    status_code=status.HTTP_201_CREATED,
    response_model=None,
)
async def create_notification_endpoint(
    body: CreateNotificationRequest,
    x_calling_domain: Annotated[
        str | None,
        Header(alias="X-Calling-Domain", convert_underscores=False),
    ] = None,
) -> dict[str, object] | JSONResponse:
    try:
        calling_domain = _normalize_x_calling_domain(x_calling_domain)
    except ValueError as exc:
        return error_response(
            "VALIDATION_ERROR",
            str(exc),
            status.HTTP_422_UNPROCESSABLE_CONTENT,
            {},
        )

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
    if calling_domain is not None:
        payload["calling_domain"] = calling_domain

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
    pipe(
        nid,
        "http.create_notification",
        message_id=n.message_id,
        state=n.state,
        issue_server_otp=body.issue_server_otp,
        calling_domain=calling_domain,
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


@app.get("/notifications/kpis")
async def get_sms_kpis_endpoint(
    created_from_raw: str | None = Query(
        None,
        alias="from",
        description=(
            "ISO 8601 inclusive lower bound on notification created_at "
            "(UTC; naive datetimes are interpreted as UTC)."
        ),
    ),
    created_to_raw: str | None = Query(
        None,
        alias="to",
        description="ISO 8601 inclusive upper bound on notification created_at (UTC).",
    ),
) -> dict[str, object]:
    """Aggregate SMS cost/volume/success KPIs (User Story 5); reads in-memory notification store."""
    dt_from = None
    dt_to = None
    try:
        if created_from_raw is not None:
            dt_from = parse_iso_datetime(created_from_raw)
        if created_to_raw is not None:
            dt_to = parse_iso_datetime(created_to_raw)
    except ValueError:
        return error_response(
            "VALIDATION_ERROR",
            "Invalid ISO 8601 datetime for from or to query parameter",
            status.HTTP_422_UNPROCESSABLE_CONTENT,
            {},
        )
    if dt_from is not None and dt_to is not None and dt_from > dt_to:
        return error_response(
            "VALIDATION_ERROR",
            "Query parameter from must be less than or equal to to",
            status.HTTP_422_UNPROCESSABLE_CONTENT,
            {},
        )
    return success_response(build_sms_kpis(created_from=dt_from, created_to=dt_to))


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
    pipe(
        notification_id,
        "dispatch.begin",
        message_id=n.message_id,
        carrier=carrier,
        as_of=body.as_of,
    )
    http_err, response = await orchestrate_select_charge_publish(
        n,
        notification_id=notification_id,
        carrier=carrier,
        as_of_ts=as_of_ts,
        flow="dispatch",
    )
    if http_err is not None:
        return http_err
    assert response is not None

    from_new = n.state
    n.state = "Send-to-provider"
    n.selected_provider_id = response.selected_provider_id
    n.routing_rule_version = int(response.routing_rule_version)
    n.updated_at = datetime.now(timezone.utc)
    update_notification(n)
    record_transition(
        n.notification_id,
        from_new,
        "Send-to-provider",
        "system",
        "accepted",
        "dispatch_new_to_send_to_provider",
    )
    pipe(
        notification_id,
        "state.Send-to-provider",
        from_state=from_new,
        estimated_cost=n.estimated_cost,
        charging_estimate_id=n.charging_estimate_id,
        currency=n.estimated_currency,
    )

    n.attempt = 1
    from_sp = n.state
    n.state = "Queue"
    n.updated_at = datetime.now(timezone.utc)
    update_notification(n)
    record_transition(
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
    pipe(
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
    pipe(
        notification_id,
        "retry.begin",
        carrier=carrier,
        current_state=n.state,
        attempt_before=n.attempt,
    )

    http_err, response = await orchestrate_select_charge_publish(
        n,
        notification_id=notification_id,
        carrier=carrier,
        as_of_ts=as_of_ts,
        flow="retry",
    )
    if http_err is not None:
        return http_err
    assert response is not None

    from_state = n.state
    n.state = "Send-to-provider"
    n.attempt += 1
    n.selected_provider_id = response.selected_provider_id
    n.routing_rule_version = int(response.routing_rule_version)
    n.updated_at = datetime.now(timezone.utc)
    update_notification(n)
    record_transition(
        n.notification_id,
        from_state,
        "Send-to-provider",
        "retry",
        "accepted",
        "retry_to_send_to_provider",
    )
    pipe(
        notification_id,
        "state.Send-to-provider",
        from_state=from_state,
        estimated_cost=n.estimated_cost,
        charging_estimate_id=n.charging_estimate_id,
        currency=n.estimated_currency,
    )

    from_sp = n.state
    n.state = "Queue"
    n.updated_at = datetime.now(timezone.utc)
    update_notification(n)
    record_transition(
        n.notification_id,
        from_sp,
        "Queue",
        "system",
        "accepted",
        "retry_send_to_provider_to_queue",
    )
    pipe(
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
    record_transition(
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
