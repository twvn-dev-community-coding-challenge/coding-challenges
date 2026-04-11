"""gRPC server factory with automatic reflection registration."""

from __future__ import annotations

from collections.abc import AsyncGenerator, Callable, Sequence
from contextlib import asynccontextmanager
from typing import Any

from google.protobuf import descriptor as _descriptor
from grpc import aio
from grpc_reflection.v1alpha import reflection


async def create_grpc_server(
    *,
    servicers: Sequence[tuple[Callable[..., None], Any]],
    descriptors: Sequence[_descriptor.FileDescriptor],
    bind: str,
) -> aio.Server:
    """Create, configure, and start an insecure gRPC server.

    Args:
        servicers: Pairs of ``(add_*Servicer_to_server, servicer_instance)``.
        descriptors: Proto ``DESCRIPTOR`` objects for reflection discovery.
        bind: Address to bind (e.g. ``"0.0.0.0:50051"``).
    """
    server = aio.server()
    for add_fn, servicer in servicers:
        add_fn(servicer, server)
    service_names = [
        svc.full_name
        for fd in descriptors
        for svc in fd.services_by_name.values()
    ]
    reflection.enable_server_reflection(
        [*service_names, reflection.SERVICE_NAME],
        server,
    )
    server.add_insecure_port(bind)
    await server.start()
    return server


def grpc_lifespan(
    *,
    servicers: Sequence[tuple[Callable[..., None], Any]],
    descriptors: Sequence[_descriptor.FileDescriptor],
    bind: str,
    grace: float = 5.0,
) -> Callable[[Any], AsyncGenerator[None, None]]:
    """Return a FastAPI lifespan that manages a gRPC server.

    The gRPC server starts on app startup and stops gracefully on shutdown.
    """

    @asynccontextmanager
    async def _lifespan(_app: Any) -> AsyncGenerator[None, None]:
        server = await create_grpc_server(
            servicers=servicers,
            descriptors=descriptors,
            bind=bind,
        )
        yield
        await server.stop(grace)

    return _lifespan
