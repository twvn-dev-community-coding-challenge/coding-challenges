"""Tests for the shared gRPC server factory."""

from __future__ import annotations

import asyncio
import logging
import socket

import pytest
from grpc import aio
from grpc_reflection.v1alpha import reflection_pb2, reflection_pb2_grpc

from generated import charging_pb2, charging_pb2_grpc, provider_pb2, provider_pb2_grpc
from py_core.redact import DEFAULT_POLICY, RedactionPolicy, RedactionRule
from py_core.server import create_grpc_server, grpc_lifespan


class _NoOpChargingServicer(charging_pb2_grpc.ChargingServiceServicer):
    pass


class _EchoEstimateChargingServicer(charging_pb2_grpc.ChargingServiceServicer):
    async def EstimateCost(
        self,
        request: charging_pb2.EstimateCostRequest,
        context: object,
    ) -> charging_pb2.EstimateCostResponse:
        return charging_pb2.EstimateCostResponse(
            estimate_id="e1",
            estimated_cost=1.0,
            currency="USD",
            rate_id="r1",
            rate_version=1,
        )


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


def test_create_grpc_server_with_tracing_enabled() -> None:
    port = _free_port()

    async def _run() -> list[str]:
        server = await create_grpc_server(
            servicers=[
                (charging_pb2_grpc.add_ChargingServiceServicer_to_server, _NoOpChargingServicer()),
            ],
            descriptors=[charging_pb2.DESCRIPTOR],
            bind=f"localhost:{port}",
            enable_tracing=True,
        )
        try:
            channel = aio.insecure_channel(f"localhost:{port}")
            return await _list_services(channel)
        finally:
            await server.stop(0)

    services = asyncio.run(_run())
    assert "gondo.charging.ChargingService" in services


def test_create_grpc_server_payload_logging(caplog: pytest.LogCaptureFixture) -> None:
    port = _free_port()
    carrier_policy = RedactionPolicy(
        rules=(
            *DEFAULT_POLICY.rules,
            RedactionRule(field_pattern="carrier", strategy="mask_phone"),
        ),
    )

    async def _run() -> None:
        server = await create_grpc_server(
            servicers=[
                (
                    charging_pb2_grpc.add_ChargingServiceServicer_to_server,
                    _EchoEstimateChargingServicer(),
                ),
            ],
            descriptors=[charging_pb2.DESCRIPTOR],
            bind=f"localhost:{port}",
            enable_payload_logging=True,
            payload_redaction_policy=carrier_policy,
        )
        try:
            channel = aio.insecure_channel(f"localhost:{port}")
            stub = charging_pb2_grpc.ChargingServiceStub(channel)
            req = charging_pb2.EstimateCostRequest(
                message_id="m1",
                provider_id="p1",
                country_code="VN",
                carrier="+84912345678",
            )
            with caplog.at_level(logging.DEBUG, logger="py_core.payload.grpc"):
                await stub.EstimateCost(req)
        finally:
            await server.stop(0)

    asyncio.run(_run())

    req_logs = [
        r
        for r in caplog.records
        if r.name == "py_core.payload.grpc" and "grpc_server_request" in r.getMessage()
    ]
    assert len(req_logs) >= 1
    body = getattr(req_logs[0], "body", None)
    assert body is not None
    assert "+84912345678" not in str(body)

    resp_logs = [
        r
        for r in caplog.records
        if r.name == "py_core.payload.grpc" and "grpc_server_response" in r.getMessage()
    ]
    assert len(resp_logs) >= 1


def test_create_grpc_server_tracing_disabled() -> None:
    port = _free_port()

    async def _run() -> list[str]:
        server = await create_grpc_server(
            servicers=[
                (charging_pb2_grpc.add_ChargingServiceServicer_to_server, _NoOpChargingServicer()),
            ],
            descriptors=[charging_pb2.DESCRIPTOR],
            bind=f"localhost:{port}",
            enable_tracing=False,
        )
        try:
            channel = aio.insecure_channel(f"localhost:{port}")
            return await _list_services(channel)
        finally:
            await server.stop(0)

    services = asyncio.run(_run())
    assert "gondo.charging.ChargingService" in services
