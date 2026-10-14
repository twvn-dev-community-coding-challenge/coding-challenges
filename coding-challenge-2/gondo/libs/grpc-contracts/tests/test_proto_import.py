"""Verify generated stubs can be imported."""

from __future__ import annotations

from generated import charging_pb2, provider_pb2
from generated import charging_pb2_grpc, provider_pb2_grpc


def test_import_pb2_modules() -> None:
    assert provider_pb2.ResolveRoutingRequest.DESCRIPTOR.full_name == (
        "gondo.provider.ResolveRoutingRequest"
    )
    assert charging_pb2.EstimateCostRequest.DESCRIPTOR.full_name == (
        "gondo.charging.EstimateCostRequest"
    )


def test_import_grpc_modules() -> None:
    assert hasattr(provider_pb2_grpc, "ProviderServiceServicer")
    assert hasattr(provider_pb2_grpc, "ProviderServiceStub")
    assert hasattr(charging_pb2_grpc, "ChargingServiceServicer")
    assert hasattr(charging_pb2_grpc, "ChargingServiceStub")
