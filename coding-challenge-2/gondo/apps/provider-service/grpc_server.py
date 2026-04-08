"""gRPC server for provider-service (placeholder servicer)."""

from __future__ import annotations

import sys
from pathlib import Path

_CONTRACTS = Path(__file__).resolve().parents[2] / "libs" / "grpc-contracts"
if str(_CONTRACTS) not in sys.path:
    sys.path.insert(0, str(_CONTRACTS))

from grpc import aio  # noqa: E402

from generated import provider_pb2, provider_pb2_grpc  # noqa: E402

PROVIDER_GRPC_BIND = "0.0.0.0:50051"


class ProviderGrpcServicer(provider_pb2_grpc.ProviderServiceServicer):
    async def ResolveRouting(
        self,
        request: provider_pb2.ResolveRoutingRequest,
        context: aio.ServicerContext,
    ) -> provider_pb2.ResolveRoutingResponse:
        candidate = provider_pb2.RoutingCandidate(
            provider_id="mock-provider-1",
            provider_code="MOCK",
            routing_rule_id="rule-mock-1",
            routing_rule_version=1,
        )
        return provider_pb2.ResolveRoutingResponse(candidates=[candidate])

    async def SelectProvider(
        self,
        request: provider_pb2.SelectProviderRequest,
        context: aio.ServicerContext,
    ) -> provider_pb2.SelectProviderResponse:
        policy = (
            request.policy_context.policy
            if request.HasField("policy_context")
            else "default"
        )
        return provider_pb2.SelectProviderResponse(
            selected_provider_id="mock-provider-1",
            selected_provider_code="MOCK",
            selection_policy=policy,
            selection_reason="placeholder_selection",
            routing_rule_id="rule-mock-1",
            routing_rule_version=1,
        )


async def create_and_start_provider_grpc_server() -> aio.Server:
    server = aio.server()
    provider_pb2_grpc.add_ProviderServiceServicer_to_server(ProviderGrpcServicer(), server)
    server.add_insecure_port(PROVIDER_GRPC_BIND)
    await server.start()
    return server
