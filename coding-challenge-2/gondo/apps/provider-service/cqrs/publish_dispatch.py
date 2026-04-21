"""CQRS write-side: publish ``sms.dispatch.requested`` after provider selection."""

from __future__ import annotations

import logging
import os
import uuid

from py_core.bus.contract import publish_json
from py_core.bus.topics import SMS_DISPATCH_REQUESTED

from bus_state import get_message_bus
from cqrs.events import SmsDispatchRequested
from yaml_registry import load_all_provider_configs

logger = logging.getLogger(__name__)


def _registry_row(provider_id: str):
    for cfg in load_all_provider_configs():
        if cfg.provider_id == provider_id:
            return cfg
    return None


def _effective_api_endpoint(selected_provider_id: str) -> str:
    """YAML ``api_endpoint`` unless ``CARRIER_HTTP_PROBE_URL`` is set (e.g. WireMock in Docker)."""
    override = os.environ.get("CARRIER_HTTP_PROBE_URL", "").strip()
    if override:
        return override
    row = _registry_row(selected_provider_id)
    return (row.api_endpoint if row and row.api_endpoint else "") or "https://sms.example.invalid"


async def publish_sms_dispatch_requested(
    *,
    message_id: str,
    country_code: str,
    carrier: str,
    selected_provider_id: str,
    selected_provider_code: str,
    estimated_cost: float | None = None,
    currency: str | None = None,
    charging_estimate_id: str | None = None,
    routing_rule_version: int | None = None,
) -> bool:
    """Publish integration event; return True on success, False if bus is not connected."""
    bus = get_message_bus()
    if bus is None:
        logger.warning("sms_dispatch_requested_skipped_no_bus")
        return False
    endpoint = _effective_api_endpoint(selected_provider_id)
    correlation_id = str(uuid.uuid4())
    event = SmsDispatchRequested(
        message_id=message_id,
        correlation_id=correlation_id,
        country_code=country_code.strip().upper(),
        carrier=carrier.strip(),
        provider_id=selected_provider_id,
        provider_code=selected_provider_code,
        api_endpoint=endpoint or "https://sms.example.invalid",
        estimated_cost=estimated_cost,
        currency=currency,
        charging_estimate_id=charging_estimate_id,
        routing_rule_version=routing_rule_version,
    )
    logger.info(
        "bus_publish_begin",
        extra={
            "topic": SMS_DISPATCH_REQUESTED,
            "message_id": message_id,
            "correlation_id": correlation_id,
            "provider_id": selected_provider_id,
        },
    )
    await publish_json(bus, SMS_DISPATCH_REQUESTED, event)
    logger.info(
        "sms_dispatch_requested_published",
        extra={
            "message_id": message_id,
            "correlation_id": correlation_id,
            "provider_id": selected_provider_id,
        },
    )
    return True
