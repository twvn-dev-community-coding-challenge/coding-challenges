"""gRPC client for provider-service → charging (EstimateCost)."""

from __future__ import annotations

import os
import sys
from pathlib import Path

_CONTRACTS = Path(__file__).resolve().parents[2] / "libs" / "grpc-contracts"
if str(_CONTRACTS) not in sys.path:
    sys.path.insert(0, str(_CONTRACTS))

from google.protobuf import timestamp_pb2  # noqa: E402
from grpc import aio  # noqa: E402

from generated import charging_pb2, charging_pb2_grpc  # noqa: E402

CHARGING_GRPC_TARGET = os.environ.get("CHARGING_GRPC_TARGET", "localhost:50052")


def _current_as_of() -> timestamp_pb2.Timestamp:
    ts = timestamp_pb2.Timestamp()
    ts.GetCurrentTime()
    return ts


async def estimate_cost_via_grpc() -> charging_pb2.EstimateCostResponse:
    async with aio.insecure_channel(CHARGING_GRPC_TARGET) as channel:
        stub = charging_pb2_grpc.ChargingServiceStub(channel)
        request = charging_pb2.EstimateCostRequest(
            message_id="msg-mock-provider-1",
            provider_id="mock-provider-1",
            country_code="US",
            carrier="mock-carrier",
            as_of=_current_as_of(),
        )
        return await stub.EstimateCost(request)
