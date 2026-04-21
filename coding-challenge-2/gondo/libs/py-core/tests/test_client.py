"""Tests for the shared gRPC client."""

from __future__ import annotations

import asyncio
import logging
import socket

import pytest
from grpc import aio
from grpc_reflection.v1alpha import reflection_pb2, reflection_pb2_grpc

from generated import charging_pb2, charging_pb2_grpc
from py_core.client import insecure_channel
from py_core.redact import DEFAULT_POLICY, RedactionPolicy, RedactionRule
from py_core.server import create_grpc_server
from py_core.tracing import TraceContext, set_current_trace_context


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


def test_insecure_channel_payload_logging(caplog: pytest.LogCaptureFixture) -> None:
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
            payload_redaction_policy=carrier_policy,
        )
        try:
            async with insecure_channel(
                f"localhost:{port}",
                enable_payload_logging=True,
                payload_redaction_policy=carrier_policy,
            ) as channel:
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

    client_logs = [
        r
        for r in caplog.records
        if r.name == "py_core.payload.grpc" and "grpc_client_request" in r.getMessage()
    ]
    assert len(client_logs) >= 1
    body = getattr(client_logs[0], "body", None)
    assert body is not None
    assert "+84912345678" not in str(body)


def test_insecure_channel_with_tracing_rpc_succeeds() -> None:
    port = _free_port()
    ctx = TraceContext(
        trace_id="a" * 32,
        span_id="b" * 16,
        parent_span_id=None,
    )
    set_current_trace_context(ctx)
    try:

        async def _run() -> list[str]:
            server = await create_grpc_server(
                servicers=[
                    (
                        charging_pb2_grpc.add_ChargingServiceServicer_to_server,
                        _NoOpChargingServicer(),
                    ),
                ],
                descriptors=[charging_pb2.DESCRIPTOR],
                bind=f"localhost:{port}",
            )
            try:
                async with insecure_channel(f"localhost:{port}") as channel:
                    return await _list_services(channel)
            finally:
                await server.stop(0)

        services = asyncio.run(_run())
        assert "gondo.charging.ChargingService" in services
    finally:
        set_current_trace_context(None)


def test_insecure_channel_tracing_disabled() -> None:
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
            async with insecure_channel(
                f"localhost:{port}",
                enable_tracing=False,
            ) as channel:
                return await _list_services(channel)
        finally:
            await server.stop(0)

    services = asyncio.run(_run())
    assert "gondo.charging.ChargingService" in services
