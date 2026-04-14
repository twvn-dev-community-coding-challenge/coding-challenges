"""Shared types for provider registry configs (YAML + gRPC mapping)."""

from __future__ import annotations

from dataclasses import dataclass, field


@dataclass(frozen=True)
class ConnectableCarrier:
    country_code: str
    carrier_code: str


@dataclass
class ProviderRegistryConfig:
    """Provider-level SMS config plus per-country carrier allowlist."""

    provider_id: str
    provider_code: str
    display_name: str
    api_endpoint: str | None
    supported_policies: list[str]
    service_status: str
    extra_config_json: str
    connectable_carriers: list[ConnectableCarrier] = field(default_factory=list)
    # Logical id for Secrets Manager / mock store (e.g. provider/prv_01).
    credentials_ref: str | None = None
