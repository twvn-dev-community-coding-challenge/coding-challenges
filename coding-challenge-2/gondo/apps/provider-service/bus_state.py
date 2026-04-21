"""Process-wide message bus handle (set during FastAPI lifespan)."""

from __future__ import annotations

from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from py_core.bus.contract import MessageBus

_bus: MessageBus | None = None


def get_message_bus() -> MessageBus | None:
    return _bus


def set_message_bus(bus: MessageBus | None) -> None:
    global _bus
    _bus = bus
