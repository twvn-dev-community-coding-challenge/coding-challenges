"""NATS subscription for inbound dispatch work."""

from __future__ import annotations

import logging
from collections.abc import AsyncGenerator
from contextlib import asynccontextmanager
from typing import Any

from cqrs.dispatch_handler import handle_dispatch_requested
from py_core.bus.factory import create_message_bus_from_env
from py_core.bus.topics import SMS_DISPATCH_REQUESTED

from bus_state import set_message_bus

logger = logging.getLogger(__name__)


@asynccontextmanager
async def carrier_lifespan(_app: Any) -> AsyncGenerator[None, None]:
    bus = create_message_bus_from_env()
    await bus.connect()
    set_message_bus(bus)
    await bus.subscribe(SMS_DISPATCH_REQUESTED, handle_dispatch_requested)
    logger.info("carrier_subscribed", extra={"topic": SMS_DISPATCH_REQUESTED})
    yield
    set_message_bus(None)
    await bus.close()
    logger.info("carrier_bus_closed")
