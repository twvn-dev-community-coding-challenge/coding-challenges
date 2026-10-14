"""Charging gRPC servicer tests."""

from __future__ import annotations

from datetime import datetime, timezone

from google.protobuf import timestamp_pb2

from generated import charging_pb2
from grpc_server import ChargingGrpcServicer


class FakeServicerContext:
    def __init__(self) -> None:
        self._code: object | None = None
        self._details: str | None = None

    async def abort(self, code: object, details: str) -> None:
        self._code = code
        self._details = details
        msg = f"gRPC abort: {code} {details}"
        raise RuntimeError(msg)


def test_charging_grpc_servicer_instantiates() -> None:
    servicer = ChargingGrpcServicer()
    assert servicer is not None


async def _estimate_vn_viettel_prv01() -> charging_pb2.EstimateCostResponse:
    servicer = ChargingGrpcServicer()
    ts = timestamp_pb2.Timestamp()
    # Before US3 cutover prv_01 still carries VN/VIETTEL in seeded rates.
    ts.FromDatetime(datetime(2026, 2, 15, 12, 0, 0, tzinfo=timezone.utc))
    req = charging_pb2.EstimateCostRequest(
        message_id="msg-grpc-1",
        provider_id="prv_01",
        country_code="VN",
        carrier="VIETTEL",
        as_of=ts,
    )
    ctx = FakeServicerContext()
    return await servicer.EstimateCost(req, ctx)


async def test_estimate_cost_returns_result() -> None:
    resp = await _estimate_vn_viettel_prv01()
    assert resp.estimate_id
    assert resp.estimated_cost == 0.015
    assert resp.currency == "USD"
    assert resp.rate_id.startswith("rate_")
    assert resp.rate_version == 1
    assert resp.HasField("created_at")
