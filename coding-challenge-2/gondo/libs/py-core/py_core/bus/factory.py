"""Construct a MessageBus from environment (default: NATS)."""

from __future__ import annotations

import os

from py_core.bus.contract import MessageBus
from py_core.bus.nats_bus import NatsMessageBus


def create_message_bus_from_env() -> MessageBus:
    """Create the default bus implementation.

    ``NATS_URL`` — comma-separated server URLs (default ``nats://127.0.0.1:4222``).
    Future: ``MESSAGE_BUS_DRIVER=kafka`` could swap implementations.
    """
    raw = os.environ.get("NATS_URL", "nats://127.0.0.1:4222")
    servers = [s.strip() for s in raw.split(",") if s.strip()]
    if not servers:
        servers = ["nats://127.0.0.1:4222"]
    return NatsMessageBus(servers=servers)
