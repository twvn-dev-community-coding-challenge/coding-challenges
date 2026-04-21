"""Carrier bounded context: MNO registry entries (separate from provider-registry)."""

from __future__ import annotations

from dataclasses import dataclass


@dataclass(frozen=True)
class CarrierRegistryEntry:
    country_code: str
    carrier_code: str
    routing_hints: dict[str, str]
    carrier_credentials_ref: str | None
