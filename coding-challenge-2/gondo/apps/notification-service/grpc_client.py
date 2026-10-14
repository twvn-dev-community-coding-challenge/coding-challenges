"""gRPC clients for notification-service -> provider."""

from __future__ import annotations

import logging
import os

from generated import provider_pb2, provider_pb2_grpc
from py_core.client import insecure_channel
from py_core.logging import _should_log_payloads

PROVIDER_GRPC_TARGET = os.environ.get("PROVIDER_GRPC_TARGET", "localhost:50051")
_PAYLOAD_LOGGING = _should_log_payloads(None)

logger = logging.getLogger(__name__)


async def select_provider(
    country_code: str,
    carrier: str,
    as_of: object,
    policy: str,
    message_id: str,
) -> provider_pb2.SelectProviderResponse:
    """Call ProviderService.SelectProvider over gRPC."""
    logger.info(
        "grpc_client_call_begin",
        extra={
            "grpc_method": "SelectProvider",
            "target": PROVIDER_GRPC_TARGET,
            "country_code": country_code,
            "carrier": carrier,
            "policy": policy,
            "message_id": message_id,
        },
    )
    async with insecure_channel(
        PROVIDER_GRPC_TARGET, enable_payload_logging=_PAYLOAD_LOGGING
    ) as channel:
        stub = provider_pb2_grpc.ProviderServiceStub(channel)
        request = provider_pb2.SelectProviderRequest(
            country_code=country_code,
            carrier=carrier,
            as_of=as_of,
            policy_context=provider_pb2.PolicyContext(
                policy=policy,
                message_id=message_id,
            ),
        )
        resp = await stub.SelectProvider(request)
    logger.info(
        "grpc_client_call_ok",
        extra={
            "grpc_method": "SelectProvider",
            "message_id": message_id,
            "selected_provider_id": resp.selected_provider_id,
            "routing_rule_version": resp.routing_rule_version,
        },
    )
    return resp


async def publish_sms_dispatch_via_provider(
    *,
    message_id: str,
    country_code: str,
    carrier: str,
    selected_provider_id: str,
    selected_provider_code: str,
    routing_rule_version: int,
    estimated_cost: float,
    currency: str,
    charging_estimate_id: str,
) -> provider_pb2.PublishSmsDispatchRequestedResponse:
    """Ask provider-service to publish ``sms.dispatch.requested`` after charging estimates."""
    logger.info(
        "grpc_client_call_begin",
        extra={
            "grpc_method": "PublishSmsDispatchRequested",
            "target": PROVIDER_GRPC_TARGET,
            "message_id": message_id,
            "country_code": country_code,
            "carrier": carrier,
            "selected_provider_id": selected_provider_id,
            "routing_rule_version": routing_rule_version,
            "charging_estimate_id": charging_estimate_id,
        },
    )
    async with insecure_channel(
        PROVIDER_GRPC_TARGET, enable_payload_logging=_PAYLOAD_LOGGING
    ) as channel:
        stub = provider_pb2_grpc.ProviderServiceStub(channel)
        request = provider_pb2.PublishSmsDispatchRequestedRequest(
            message_id=message_id,
            country_code=country_code,
            carrier=carrier,
            selected_provider_id=selected_provider_id,
            selected_provider_code=selected_provider_code,
            routing_rule_version=routing_rule_version,
            estimated_cost=estimated_cost,
            currency=currency,
            charging_estimate_id=charging_estimate_id,
        )
        resp = await stub.PublishSmsDispatchRequested(request)
    logger.info(
        "grpc_client_call_ok",
        extra={
            "grpc_method": "PublishSmsDispatchRequested",
            "message_id": message_id,
            "published": resp.published,
        },
    )
    return resp


async def resolve_routing(
    country_code: str,
    carrier: str,
    as_of: object,
) -> provider_pb2.ResolveRoutingResponse:
    """Call ProviderService.ResolveRouting over gRPC."""
    logger.info(
        "grpc_client_call_begin",
        extra={
            "grpc_method": "ResolveRouting",
            "target": PROVIDER_GRPC_TARGET,
            "country_code": country_code,
            "carrier": carrier,
        },
    )
    async with insecure_channel(
        PROVIDER_GRPC_TARGET, enable_payload_logging=_PAYLOAD_LOGGING
    ) as channel:
        stub = provider_pb2_grpc.ProviderServiceStub(channel)
        request = provider_pb2.ResolveRoutingRequest(
            country_code=country_code,
            carrier=carrier,
            as_of=as_of,
        )
        resp = await stub.ResolveRouting(request)
    logger.info(
        "grpc_client_call_ok",
        extra={
            "grpc_method": "ResolveRouting",
            "candidate_count": len(resp.candidates),
        },
    )
    return resp
