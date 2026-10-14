"""Tests for the shared gRPC reflection plugin."""

from __future__ import annotations

import asyncio

from grpc import aio
from grpc_reflection.v1alpha import reflection_pb2, reflection_pb2_grpc

from generated import charging_pb2, charging_pb2_grpc, provider_pb2, provider_pb2_grpc
from reflection import enable_reflection


class _NoOpChargingServicer(charging_pb2_grpc.ChargingServiceServicer):
    pass


class _NoOpProviderServicer(provider_pb2_grpc.ProviderServiceServicer):
    pass


async def _list_services(channel: aio.Channel) -> list[str]:
    stub = reflection_pb2_grpc.ServerReflectionStub(channel)
    request = reflection_pb2.ServerReflectionRequest(list_services="")
    responses: list[reflection_pb2.ServerReflectionResponse] = []
    async for resp in stub.ServerReflectionInfo(iter([request])):
        responses.append(resp)
    names = [
        svc.name
        for resp in responses
        for svc in resp.list_services_response.service
    ]
    return sorted(names)


def test_enable_reflection_registers_service_names() -> None:
    async def _run() -> list[str]:
        server = aio.server()
        charging_pb2_grpc.add_ChargingServiceServicer_to_server(
            _NoOpChargingServicer(), server
        )
        enable_reflection(server, [charging_pb2.DESCRIPTOR])
        port = server.add_insecure_port("localhost:0")
        await server.start()
        try:
            channel = aio.insecure_channel(f"localhost:{port}")
            return await _list_services(channel)
        finally:
            await server.stop(0)

    services = asyncio.get_event_loop().run_until_complete(_run())
    assert "gondo.charging.ChargingService" in services
    assert "grpc.reflection.v1alpha.ServerReflection" in services


def test_enable_reflection_multiple_descriptors() -> None:
    async def _run() -> list[str]:
        server = aio.server()
        charging_pb2_grpc.add_ChargingServiceServicer_to_server(
            _NoOpChargingServicer(), server
        )
        provider_pb2_grpc.add_ProviderServiceServicer_to_server(
            _NoOpProviderServicer(), server
        )
        enable_reflection(
            server, [charging_pb2.DESCRIPTOR, provider_pb2.DESCRIPTOR]
        )
        port = server.add_insecure_port("localhost:0")
        await server.start()
        try:
            channel = aio.insecure_channel(f"localhost:{port}")
            return await _list_services(channel)
        finally:
            await server.stop(0)

    services = asyncio.get_event_loop().run_until_complete(_run())
    assert "gondo.charging.ChargingService" in services
    assert "gondo.provider.ProviderService" in services
    assert "grpc.reflection.v1alpha.ServerReflection" in services


def test_enable_reflection_no_descriptors_still_registers_reflection() -> None:
    async def _run() -> list[str]:
        server = aio.server()
        enable_reflection(server, [])
        port = server.add_insecure_port("localhost:0")
        await server.start()
        try:
            channel = aio.insecure_channel(f"localhost:{port}")
            return await _list_services(channel)
        finally:
            await server.stop(0)

    services = asyncio.get_event_loop().run_until_complete(_run())
    assert "grpc.reflection.v1alpha.ServerReflection" in services
