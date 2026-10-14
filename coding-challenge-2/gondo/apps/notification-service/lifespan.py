"""FastAPI lifespan: NATS subscriber for carrier dispatch-received (async messaging)."""

from __future__ import annotations

import logging
from collections.abc import AsyncGenerator
from contextlib import asynccontextmanager
from typing import Any

from py_core.bus.factory import create_message_bus_from_env

from cqrs.dispatch_received_subscriber import subscribe_to_dispatch_received

logger = logging.getLogger(__name__)


@asynccontextmanager
async def notification_lifespan(_app: Any) -> AsyncGenerator[None, None]:
    bus = create_message_bus_from_env()
    await bus.connect()
    await subscribe_to_dispatch_received(bus)
    logger.info("notification_message_bus_ready")
    try:
        yield
    finally:
        await bus.close()
        logger.info("notification_message_bus_stopped")
