"""Compose gRPC server + NATS message bus lifespan."""

from __future__ import annotations

import logging
from collections.abc import AsyncGenerator, Callable
from contextlib import asynccontextmanager
from typing import Any

from generated import provider_pb2, provider_pb2_grpc
from grpc_server import ProviderGrpcServicer
from py_core.bus.factory import create_message_bus_from_env
from py_core.server import grpc_lifespan

from bus_state import set_message_bus
from cqrs.outcome_subscriber import subscribe_to_dispatch_outcomes

logger = logging.getLogger(__name__)


def build_provider_lifespan(
    *,
    enable_payload_logging: bool | None,
) -> Callable[[Any], AsyncGenerator[None, None]]:
    """gRPC + message bus: publish/consume via NATS (sidecar broker)."""

    grpc_cm = grpc_lifespan(
        servicers=[
            (
                provider_pb2_grpc.add_ProviderServiceServicer_to_server,
                ProviderGrpcServicer(),
            ),
        ],
        descriptors=[provider_pb2.DESCRIPTOR],
        bind="0.0.0.0:50051",
        enable_payload_logging=enable_payload_logging,
    )

    @asynccontextmanager
    async def _lifespan(app: Any) -> AsyncGenerator[None, None]:
        bus = create_message_bus_from_env()
        await bus.connect()
        set_message_bus(bus)
        await subscribe_to_dispatch_outcomes(bus)
        logger.info("provider_message_bus_ready")
        async with grpc_cm(app):
            yield
        set_message_bus(None)
        await bus.close()
        logger.info("provider_message_bus_stopped")

    return _lifespan
