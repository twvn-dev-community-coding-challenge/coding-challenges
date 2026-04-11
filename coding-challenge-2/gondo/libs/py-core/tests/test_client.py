"""Tests for the shared gRPC client."""

from __future__ import annotations

import asyncio
import socket

from grpc import aio
from grpc_reflection.v1alpha import reflection_pb2, reflection_pb2_grpc

from generated import charging_pb2, charging_pb2_grpc
from py_core.client import insecure_channel
from py_core.server import create_grpc_server


class _NoOpChargingServicer(charging_pb2_grpc.ChargingServiceServicer):
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


def test_insecure_channel_reflection_rpc_succeeds() -> None:
    port = _free_port()

    async def _run() -> tuple[list[str], bool]:
        server = await create_grpc_server(
            servicers=[
                (charging_pb2_grpc.add_ChargingServiceServicer_to_server, _NoOpChargingServicer()),
            ],
            descriptors=[charging_pb2.DESCRIPTOR],
            bind=f"localhost:{port}",
        )
        try:
            ch: aio.Channel | None = None
            services: list[str] = []
            async with insecure_channel(f"localhost:{port}") as channel:
                ch = channel
                services = await _list_services(channel)
            use_after_close_failed = False
            if ch is not None:
                try:
                    await _list_services(ch)
                except Exception:
                    use_after_close_failed = True
            return services, use_after_close_failed
        finally:
            await server.stop(0)

    services, closed_prevents_use = asyncio.run(_run())
    assert "gondo.charging.ChargingService" in services
    assert closed_prevents_use is True
