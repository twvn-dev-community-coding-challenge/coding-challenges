"""Charging gRPC servicer can be constructed."""

from __future__ import annotations

from grpc_server import ChargingGrpcServicer


def test_charging_grpc_servicer_instantiates() -> None:
    servicer = ChargingGrpcServicer()
    assert servicer is not None
