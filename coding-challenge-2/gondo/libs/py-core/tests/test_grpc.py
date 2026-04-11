"""Tests for the shared gRPC server factory."""

from __future__ import annotations

import asyncio
import socket

from grpc import aio
from grpc_reflection.v1alpha import reflection_pb2, reflection_pb2_grpc

from generated import charging_pb2, charging_pb2_grpc, provider_pb2, provider_pb2_grpc
from py_core.server import create_grpc_server, grpc_lifespan


class _NoOpChargingServicer(charging_pb2_grpc.ChargingServiceServicer):
    pass


class _NoOpProviderServicer(provider_pb2_grpc.ProviderServiceServicer):
    pass


def _free_port() -> int:
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        s.bind(("localhost", 0))
        return s.getsockname()[1]


async def _list_services(channel: aio.Channel) -> list[str]:
    stub = reflection_pb2_grpc.ServerReflectionStub(channel)
    request = reflection_pb2.ServerReflectionRequest(list_services="")
    responses: list[reflection_pb2.ServerReflectionResponse] = []
    async for resp in stub.ServerReflectionInfo(iter([request])):
        responses.append(resp)
    return sorted(svc.name for resp in responses for svc in resp.list_services_response.service)


def test_create_grpc_server_registers_servicer_and_reflection() -> None:
    port = _free_port()

    async def _run() -> list[str]:
        server = await create_grpc_server(
            servicers=[
                (charging_pb2_grpc.add_ChargingServiceServicer_to_server, _NoOpChargingServicer()),
            ],
            descriptors=[charging_pb2.DESCRIPTOR],
            bind=f"localhost:{port}",
        )
        try:
            channel = aio.insecure_channel(f"localhost:{port}")
            return await _list_services(channel)
        finally:
            await server.stop(0)

    services = asyncio.run(_run())
    assert "gondo.charging.ChargingService" in services
    assert "grpc.reflection.v1alpha.ServerReflection" in services


def test_create_grpc_server_multiple_servicers() -> None:
    port = _free_port()

    async def _run() -> list[str]:
        server = await create_grpc_server(
            servicers=[
                (charging_pb2_grpc.add_ChargingServiceServicer_to_server, _NoOpChargingServicer()),
                (provider_pb2_grpc.add_ProviderServiceServicer_to_server, _NoOpProviderServicer()),
            ],
            descriptors=[charging_pb2.DESCRIPTOR, provider_pb2.DESCRIPTOR],
            bind=f"localhost:{port}",
        )
        try:
            channel = aio.insecure_channel(f"localhost:{port}")
            return await _list_services(channel)
        finally:
            await server.stop(0)

    services = asyncio.run(_run())
    assert "gondo.charging.ChargingService" in services
    assert "gondo.provider.ProviderService" in services


def test_grpc_lifespan_starts_and_stops_server() -> None:
    port = _free_port()

    async def _run() -> bool:
        lifespan = grpc_lifespan(
            servicers=[
                (charging_pb2_grpc.add_ChargingServiceServicer_to_server, _NoOpChargingServicer()),
            ],
            descriptors=[charging_pb2.DESCRIPTOR],
            bind=f"localhost:{port}",
        )
        started = False
        async with lifespan(None):
            started = True
        return started

    result = asyncio.run(_run())
    assert result is True
