"""gRPC servicer for provider-service."""

from __future__ import annotations

import grpc
from grpc import aio

from generated import provider_pb2, provider_pb2_grpc
from py_core.proto_utils import as_utc, current_timestamp, datetime_to_timestamp
from registry import (
    RoutingCandidate as DomainRoutingCandidate,
    resolve_routing,
    select_provider,
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
        result = await select_provider(
            request.country_code,
            request.carrier,
            as_of,
            policy,
            message_id=message_id,
        )
        if result is None:
            await context.abort(grpc.StatusCode.NOT_FOUND, "ROUTING_UNRESOLVABLE")
        return provider_pb2.SelectProviderResponse(
            selected_provider_id=result.selected_provider_id,
            selected_provider_code=result.selected_provider_code,
            selection_policy=result.selection_policy,
            selection_reason=result.selection_reason,
            routing_rule_id=result.routing_rule_id,
            routing_rule_version=result.routing_rule_version,
            selected_at=current_timestamp(),
        )
