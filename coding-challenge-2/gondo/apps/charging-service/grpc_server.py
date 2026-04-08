"""gRPC server for charging-service (placeholder servicer)."""

from __future__ import annotations

import sys
from pathlib import Path

_CONTRACTS = Path(__file__).resolve().parents[2] / "libs" / "grpc-contracts"
if str(_CONTRACTS) not in sys.path:
    sys.path.insert(0, str(_CONTRACTS))

from grpc import aio  # noqa: E402

from generated import charging_pb2, charging_pb2_grpc  # noqa: E402

CHARGING_GRPC_BIND = "0.0.0.0:50052"


class ChargingGrpcServicer(charging_pb2_grpc.ChargingServiceServicer):
    async def EstimateCost(
        self,
        request: charging_pb2.EstimateCostRequest,
        context: aio.ServicerContext,
    ) -> charging_pb2.EstimateCostResponse:
        return charging_pb2.EstimateCostResponse(
            estimate_id="est-mock-1",
            estimated_cost=0.0123,
            currency="USD",
            rate_id="rate-mock-1",
            rate_version=1,
        )

    async def RecordActualCost(
        self,
        request: charging_pb2.RecordActualCostRequest,
        context: aio.ServicerContext,
    ) -> charging_pb2.RecordActualCostResponse:
        return charging_pb2.RecordActualCostResponse(
            actual_cost_id="actual-mock-1",
            message_id=request.message_id,
            actual_cost=request.actual_cost,
            idempotent_replay=False,
        )


async def create_and_start_charging_grpc_server() -> aio.Server:
    server = aio.server()
    charging_pb2_grpc.add_ChargingServiceServicer_to_server(ChargingGrpcServicer(), server)
    server.add_insecure_port(CHARGING_GRPC_BIND)
    await server.start()
    return server
