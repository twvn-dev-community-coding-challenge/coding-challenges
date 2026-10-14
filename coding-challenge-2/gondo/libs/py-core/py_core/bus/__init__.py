"""Broker-agnostic messaging primitives (NATS-backed default; swappable via factory)."""

from py_core.bus.contract import MessageBus, Subscription, publish_json
from py_core.bus.factory import create_message_bus_from_env
from py_core.bus.nats_bus import NatsMessageBus
from py_core.bus.topics import SMS_DISPATCH_OUTCOME, SMS_DISPATCH_REQUESTED, topic_to_subject

__all__ = [
    "MessageBus",
    "NatsMessageBus",
    "Subscription",
    "create_message_bus_from_env",
    "publish_json",
    "SMS_DISPATCH_OUTCOME",
    "SMS_DISPATCH_REQUESTED",
    "topic_to_subject",
]
