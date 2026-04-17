"""CQRS event payloads (integration / SMS dispatch pipeline)."""

from __future__ import annotations

from dataclasses import dataclass


@dataclass(frozen=True)
class SmsDispatchRequested:
    """Command side-effect: enqueue work for carrier-service."""

    message_id: str
    correlation_id: str
    country_code: str
    carrier: str
    provider_id: str
    provider_code: str
    api_endpoint: str


@dataclass(frozen=True)
class SmsDispatchOutcome:
    """Published by carrier-service; consumed by provider-service."""

    status: str  # success | failure
    message_id: str
    correlation_id: str
    detail: str
