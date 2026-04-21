"""Provider selection, charging estimate, and bus publish for dispatch/retry commands."""

from __future__ import annotations

import logging
from datetime import datetime, timezone
from typing import Literal

import grpc
from fastapi import status
from fastapi.responses import JSONResponse
from google.protobuf import timestamp_pb2

from charging_client import estimate_cost_grpc
from generated import provider_pb2
from grpc_client import publish_sms_dispatch_via_provider, select_provider
from models import Notification
from pipeline_runtime import pipe
from responses import error_response
from store import update_notification

logger = logging.getLogger(__name__)

DispatchFlowKind = Literal["dispatch", "retry"]

_FLOW_LOGGER: dict[DispatchFlowKind, str] = {
    "dispatch": "notification_dispatch_flow",
    "retry": "notification_retry_flow",
}
_EXC_SELECT: dict[DispatchFlowKind, str] = {
    "dispatch": "provider_selection_failed",
    "retry": "provider_selection_failed_retry",
}
_EXC_PUBLISH: dict[DispatchFlowKind, str] = {
    "dispatch": "publish_dispatch_requested_failed",
    "retry": "publish_dispatch_requested_failed_retry",
}


def iso_string_to_timestamp(iso: str) -> timestamp_pb2.Timestamp:
    normalized = iso.replace("Z", "+00:00")
    parsed = datetime.fromisoformat(normalized)
    if parsed.tzinfo is None:
        parsed = parsed.replace(tzinfo=timezone.utc)
    ts = timestamp_pb2.Timestamp()
    ts.FromDatetime(parsed.astimezone(timezone.utc))
    return ts


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
        pipe(
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
    pipe(
        n.notification_id,
        "grpc.EstimateCost.ok",
        charging_estimate_id=n.charging_estimate_id,
        estimated_cost=n.estimated_cost,
        currency=n.estimated_currency,
        rate_id=n.charging_rate_id,
    )
    return None


async def orchestrate_select_charge_publish(
    n: Notification,
    *,
    notification_id: str,
    carrier: str,
    as_of_ts: timestamp_pb2.Timestamp,
    flow: DispatchFlowKind,
) -> tuple[JSONResponse | None, provider_pb2.SelectProviderResponse | None]:
    """Run SelectProvider → EstimateCost → PublishSmsDispatchRequested.

    Returns ``(JSONResponse, None)`` on failure or ``(None, response)`` on success.
    """
    log_key = _FLOW_LOGGER[flow]
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
            _EXC_SELECT[flow],
            extra={"notification_id": notification_id, "details": err.details()},
        )
        pipe(
            notification_id,
            "grpc.SelectProvider.failed",
            grpc_code=str(err.code()),
            details=err.details(),
        )
        return (
            error_response(
                "PROVIDER_SELECTION_FAILED",
                "Provider selection failed",
                status.HTTP_502_BAD_GATEWAY,
                {"grpc": err.details() or str(err.code())},
            ),
            None,
        )

    logger.info(
        log_key,
        extra={
            "step": "after_select_provider",
            "notification_id": notification_id,
            "message_id": n.message_id,
            "selected_provider_id": response.selected_provider_id,
            "routing_rule_version": response.routing_rule_version,
        },
    )
    if flow == "dispatch":
        pipe(
            notification_id,
            "grpc.SelectProvider.ok",
            selected_provider_id=response.selected_provider_id,
            selected_provider_code=response.selected_provider_code,
            routing_rule_version=int(response.routing_rule_version),
        )
    else:
        pipe(
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
        return charging_err, None

    if n.message_id:
        logger.info(
            log_key,
            extra={
                "step": "before_publish_dispatch_requested",
                "notification_id": notification_id,
                "message_id": n.message_id,
            },
        )
        pipe(notification_id, f"{flow}.before_publish_dispatch_requested")
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
                _EXC_PUBLISH[flow],
                extra={"notification_id": notification_id, "details": err.details()},
            )
            pipe(
                notification_id,
                "grpc.PublishSmsDispatchRequested.failed",
                grpc_code=str(err.code()),
                details=err.details(),
            )
            return (
                error_response(
                    "DISPATCH_PUBLISH_FAILED",
                    "Could not publish sms.dispatch.requested via provider-service",
                    status.HTTP_502_BAD_GATEWAY,
                    {"grpc": err.details() or str(err.code())},
                ),
                None,
            )
        if not pub_resp.published:
            pipe(notification_id, "bus.sms_dispatch.requested.skipped", published=False)
            return (
                error_response(
                    "DISPATCH_REQUEST_NOT_PUBLISHED",
                    "sms.dispatch.requested was not published; message bus may be unavailable",
                    status.HTTP_503_SERVICE_UNAVAILABLE,
                    {"message_id": n.message_id},
                ),
                None,
            )
        logger.info(
            log_key,
            extra={
                "step": "after_publish_dispatch_requested",
                "notification_id": notification_id,
                "message_id": n.message_id,
                "published": pub_resp.published,
            },
        )
        pipe(
            notification_id,
            "grpc.PublishSmsDispatchRequested.ok",
            published=True,
            charging_estimate_id=n.charging_estimate_id,
        )
    else:
        logger.warning(
            "notification_dispatch_skip_bus_publish"
            if flow == "dispatch"
            else "notification_retry_skip_bus_publish",
            extra={
                "notification_id": notification_id,
                "reason": "empty_message_id",
            },
        )
        pipe(notification_id, f"{flow}.skip_bus_publish", reason="empty_message_id")

    return None, response
