"""Provider gRPC servicer can be constructed."""

from __future__ import annotations

from grpc_server import ProviderGrpcServicer


def test_provider_grpc_servicer_instantiates() -> None:
    servicer = ProviderGrpcServicer()
    assert servicer is not None
