"""grpc_client module loads with grpc-contracts on path (see project.json env)."""

from __future__ import annotations

import grpc_client


def test_grpc_client_constants() -> None:
    assert grpc_client.PROVIDER_GRPC_TARGET == "localhost:50051"


def test_grpc_client_callables_exported() -> None:
    assert callable(grpc_client.select_provider)
    assert callable(grpc_client.resolve_routing)
