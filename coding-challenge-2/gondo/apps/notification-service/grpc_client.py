"""gRPC clients for notification-service -> provider."""

from __future__ import annotations

import os

from generated import provider_pb2, provider_pb2_grpc
from py_core.client import insecure_channel
from py_core.logging import _should_log_payloads

PROVIDER_GRPC_TARGET = os.environ.get("PROVIDER_GRPC_TARGET", "localhost:50051")
_PAYLOAD_LOGGING = _should_log_payloads(None)


async def select_provider(
    country_code: str,
    carrier: str,
    as_of: object,
    policy: str,
    message_id: str,
) -> provider_pb2.SelectProviderResponse:
    """Call ProviderService.SelectProvider over gRPC."""
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
        return await stub.SelectProvider(request)


async def resolve_routing(
    country_code: str,
    carrier: str,
    as_of: object,
) -> provider_pb2.ResolveRoutingResponse:
    """Call ProviderService.ResolveRouting over gRPC."""
    async with insecure_channel(
        PROVIDER_GRPC_TARGET, enable_payload_logging=_PAYLOAD_LOGGING
    ) as channel:
        stub = provider_pb2_grpc.ProviderServiceStub(channel)
        request = provider_pb2.ResolveRoutingRequest(
            country_code=country_code,
            carrier=carrier,
            as_of=as_of,
        )
        return await stub.ResolveRouting(request)
