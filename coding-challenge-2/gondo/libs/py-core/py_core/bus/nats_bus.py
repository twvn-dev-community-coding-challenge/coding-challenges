"""NATS implementation of MessageBus (JetStream-ready subject layout)."""

from __future__ import annotations

import logging
from typing import TYPE_CHECKING

import nats
from nats.aio.client import Client as NATS
from nats.aio.subscription import Subscription as NatsSubscription

from py_core.bus.contract import MessageBus, MessageHandler, Subscription
from py_core.bus.topics import topic_to_subject

if TYPE_CHECKING:
    pass

logger = logging.getLogger(__name__)


class _NatsSubscription(Subscription):
    def __init__(self, inner: NatsSubscription) -> None:
        self._inner = inner

    async def unsubscribe(self) -> None:
        await self._inner.unsubscribe()


class NatsMessageBus:
    """Thin wrapper: logical topics → `gondo.<topic>` NATS subjects."""

    def __init__(self, servers: list[str]) -> None:
        self._servers = servers
        self._nc: NATS | None = None

    async def connect(self) -> None:
        if self._nc is not None:
            return
        self._nc = await nats.connect(servers=self._servers)
        logger.info("nats_connected", extra={"servers": self._servers})

    async def close(self) -> None:
        if self._nc is None:
            return
        await self._nc.drain()
        await self._nc.close()
        self._nc = None

    async def publish(self, topic: str, payload: bytes) -> None:
        if self._nc is None:
            msg = "NatsMessageBus.publish before connect"
            raise RuntimeError(msg)
        subject = topic_to_subject(topic)
        await self._nc.publish(subject, payload)

    async def subscribe(
        self,
        topic: str,
        handler: MessageHandler,
    ) -> Subscription:
        if self._nc is None:
            msg = "NatsMessageBus.subscribe before connect"
            raise RuntimeError(msg)
        subject = topic_to_subject(topic)

        async def _async_cb(msg: object) -> None:
            data = getattr(msg, "data", None)
            if not isinstance(data, (bytes, bytearray)):
                return
            payload = bytes(data)
            await handler(payload)

        sub = await self._nc.subscribe(subject, cb=_async_cb)
        return _NatsSubscription(sub)
