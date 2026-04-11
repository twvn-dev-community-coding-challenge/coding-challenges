"""gRPC client primitives (async channel helpers)."""

from __future__ import annotations

from collections.abc import AsyncIterator
from contextlib import asynccontextmanager

from grpc import aio
from grpc.aio import Channel


@asynccontextmanager
async def insecure_channel(target: str) -> AsyncIterator[Channel]:
    """Open an insecure async gRPC channel to ``target`` (``host:port``).

    Yields a :class:`grpc.aio.Channel` for use with generated stubs.
    The channel is closed when the context exits (success or failure).
    """
    channel = aio.insecure_channel(target)
    try:
        yield channel
    finally:
        await channel.close()
