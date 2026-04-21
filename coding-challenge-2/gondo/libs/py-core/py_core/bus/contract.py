"""Abstract message bus — apps depend on this, not on NATS/Kafka directly."""

from __future__ import annotations

from collections.abc import Awaitable, Callable
from typing import Protocol, runtime_checkable


@runtime_checkable
class Subscription(Protocol):
    async def unsubscribe(self) -> None: ...


MessageHandler = Callable[[bytes], Awaitable[None]]


@runtime_checkable
class MessageBus(Protocol):
    """Publish/subscribe with logical topic names (no broker-specific subjects)."""

    async def connect(self) -> None:
        """Establish connection to the broker (or no-op for HTTP gateway stubs)."""

    async def close(self) -> None:
        """Drain and disconnect."""

    async def publish(self, topic: str, payload: bytes) -> None:
        """Publish raw bytes to a logical topic."""

    async def subscribe(
        self,
        topic: str,
        handler: MessageHandler,
    ) -> Subscription:
        """Subscribe; handler invoked per message (bytes body)."""


async def publish_json(bus: MessageBus, topic: str, obj: object) -> None:
    """Serialize dict/dataclass-like object to JSON bytes."""
    import json
    from dataclasses import asdict, is_dataclass

    if is_dataclass(obj) and not isinstance(obj, type):
        body = json.dumps(asdict(obj)).encode("utf-8")
    elif isinstance(obj, dict):
        body = json.dumps(obj).encode("utf-8")
    else:
        body = json.dumps(obj, default=str).encode("utf-8")
    await bus.publish(topic, body)
