"""gRPC servicer for provider-service."""

from __future__ import annotations

import logging

import grpc
from grpc import aio

from generated import provider_pb2, provider_pb2_grpc
from py_core.proto_utils import as_utc, current_timestamp, datetime_to_timestamp
from py_core.credentials_store import backend_name, secret_configured

from cqrs.publish_dispatch import publish_sms_dispatch_requested
from registry import (
    ConnectableCarrier as DomainConnectableCarrier,
    ProviderRegistryConfig as DomainProviderRegistryConfig,
    RoutingCandidate as DomainRoutingCandidate,
    get_provider_registry,
    resolve_routing,
    select_provider,
)

logger = logging.getLogger(__name__)


def _connectable_carrier_to_proto(
    c: DomainConnectableCarrier,
) -> provider_pb2.ConnectableCarrier:
    return provider_pb2.ConnectableCarrier(
        country_code=c.country_code,
        carrier_code=c.carrier_code,
    )


def _registry_config_to_proto(
    c: DomainProviderRegistryConfig,
) -> provider_pb2.ProviderRegistryConfig:
    cref = c.credentials_ref or ""
    return provider_pb2.ProviderRegistryConfig(
        provider_id=c.provider_id,
        provider_code=c.provider_code,
        display_name=c.display_name,
        api_endpoint=c.api_endpoint or "",
        supported_policies=c.supported_policies,
        service_status=c.service_status,
        extra_config_json=c.extra_config_json,
        connectable_carriers=[
            _connectable_carrier_to_proto(x) for x in c.connectable_carriers
        ],
        credentials_ref=cref,
        credentials_configured=secret_configured(c.credentials_ref),
        credentials_backend=backend_name(),
    )


def _routing_candidate_to_proto(
    c: DomainRoutingCandidate,
) -> provider_pb2.RoutingCandidate:
    msg = provider_pb2.RoutingCandidate(
        provider_id=c.provider_id,
        provider_code=c.provider_code,
        routing_rule_id=c.routing_rule_id,
        routing_rule_version=c.routing_rule_version,
        effective_from=datetime_to_timestamp(c.effective_from),
        resolved_at=datetime_to_timestamp(c.resolved_at),
    )
    if c.effective_to is not None:
        msg.effective_to.CopyFrom(datetime_to_timestamp(c.effective_to))
    return msg


class ProviderGrpcServicer(provider_pb2_grpc.ProviderServiceServicer):
    async def ResolveRouting(
        self,
        request: provider_pb2.ResolveRoutingRequest,
        context: aio.ServicerContext,
    ) -> provider_pb2.ResolveRoutingResponse:
        if not request.HasField("as_of"):
            await context.abort(
                grpc.StatusCode.INVALID_ARGUMENT,
                "as_of is required",
            )
        as_of = as_utc(request.as_of.ToDatetime())
        candidates = await resolve_routing(request.country_code, request.carrier, as_of)
        if not candidates:
            await context.abort(grpc.StatusCode.NOT_FOUND, "ROUTING_UNRESOLVABLE")
        return provider_pb2.ResolveRoutingResponse(
            candidates=[_routing_candidate_to_proto(c) for c in candidates],
        )

    async def SelectProvider(
        self,
        request: provider_pb2.SelectProviderRequest,
        context: aio.ServicerContext,
    ) -> provider_pb2.SelectProviderResponse:
        if not request.HasField("as_of"):
            await context.abort(
                grpc.StatusCode.INVALID_ARGUMENT,
                "as_of is required",
            )
        as_of = as_utc(request.as_of.ToDatetime())
        policy = "highest_precedence"
        message_id = ""
        if request.HasField("policy_context"):
            policy = request.policy_context.policy or "highest_precedence"
            message_id = request.policy_context.message_id
        logger.info(
            "provider_grpc_request",
            extra={
                "grpc_method": "SelectProvider",
                "country_code": request.country_code,
                "carrier": request.carrier,
                "policy": policy,
                "message_id": message_id,
            },
        )
        result = await select_provider(
            request.country_code,
            request.carrier,
            as_of,
            policy,
            message_id=message_id,
        )
        if result is None:
            await context.abort(grpc.StatusCode.NOT_FOUND, "ROUTING_UNRESOLVABLE")
        logger.info(
            "provider_grpc_response",
            extra={
                "grpc_method": "SelectProvider",
                "message_id": message_id,
                "selected_provider_id": result.selected_provider_id,
                "routing_rule_version": result.routing_rule_version,
            },
        )
        return provider_pb2.SelectProviderResponse(
            selected_provider_id=result.selected_provider_id,
            selected_provider_code=result.selected_provider_code,
            selection_policy=result.selection_policy,
            selection_reason=result.selection_reason,
            routing_rule_id=result.routing_rule_id,
            routing_rule_version=result.routing_rule_version,
            selected_at=current_timestamp(),
            sms_dispatch_requested_published=False,
        )

    async def PublishSmsDispatchRequested(
        self,
        request: provider_pb2.PublishSmsDispatchRequestedRequest,
        context: aio.ServicerContext,
    ) -> provider_pb2.PublishSmsDispatchRequestedResponse:
        mid = (request.message_id or "").strip()
        logger.info(
            "provider_grpc_request",
            extra={
                "grpc_method": "PublishSmsDispatchRequested",
                "message_id": mid or request.message_id,
                "country_code": request.country_code,
                "carrier": request.carrier,
                "selected_provider_id": request.selected_provider_id,
                "routing_rule_version": request.routing_rule_version,
                "charging_estimate_id": request.charging_estimate_id,
            },
        )
        if not mid:
            await context.abort(grpc.StatusCode.INVALID_ARGUMENT, "message_id is required")
        if not (request.selected_provider_id or "").strip():
            await context.abort(
                grpc.StatusCode.INVALID_ARGUMENT,
                "selected_provider_id is required",
            )
        est = request.estimated_cost
        cur = (request.currency or "").strip() or None
        eid = (request.charging_estimate_id or "").strip() or None
        rrv = int(request.routing_rule_version)
        published = await publish_sms_dispatch_requested(
            message_id=mid,
            country_code=request.country_code,
            carrier=request.carrier,
            selected_provider_id=request.selected_provider_id,
            selected_provider_code=request.selected_provider_code,
            estimated_cost=est,
            currency=cur,
            charging_estimate_id=eid,
            routing_rule_version=rrv,
        )
        logger.info(
            "provider_grpc_response",
            extra={
                "grpc_method": "PublishSmsDispatchRequested",
                "message_id": mid,
                "published": published,
            },
        )
        return provider_pb2.PublishSmsDispatchRequestedResponse(published=published)

    async def GetProviderRegistry(
        self,
        request: provider_pb2.GetProviderRegistryRequest,
        context: aio.ServicerContext,
    ) -> provider_pb2.GetProviderRegistryResponse:
        country = request.country_code.strip().upper()
        if not country:
            await context.abort(
                grpc.StatusCode.INVALID_ARGUMENT,
                "country_code is required",
            )
        carrier = request.carrier.strip() if request.carrier else None
        provider_id = request.provider_id.strip() if request.provider_id else None
        policy = request.policy.strip() if request.policy else None
        configs = await get_provider_registry(country, carrier, provider_id, policy)
        return provider_pb2.GetProviderRegistryResponse(
            configs=[_registry_config_to_proto(c) for c in configs],
        )
