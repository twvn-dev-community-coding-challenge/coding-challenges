"""Load provider SMS metadata from the infra YAML registry (country-partitioned catalogs)."""

from __future__ import annotations

import json
import logging
import os
from functools import lru_cache
from pathlib import Path
from typing import Any

import yaml

from registry_types import ConnectableCarrier, ProviderRegistryConfig

_INDEX_NAME = "index.yaml"

logger = logging.getLogger(__name__)


def _default_registry_root() -> Path:
    base = Path(__file__).resolve().parent.parent.parent
    return base / "infra" / "provider-registry"


def registry_root() -> Path:
    override = os.environ.get("PROVIDER_REGISTRY_ROOT")
    if override:
        return Path(override).expanduser().resolve()
    return _default_registry_root()


def _read_yaml(path: Path) -> Any:
    text = path.read_text(encoding="utf-8")
    return yaml.safe_load(text)


def _parse_provider_doc(
    raw: dict[str, Any],
    source: Path,
    *,
    implicit_country: str | None = None,
) -> ProviderRegistryConfig:
    pid = raw.get("provider_id")
    pcode = raw.get("provider_code")
    name = raw.get("display_name")
    if not isinstance(pid, str) or not isinstance(pcode, str) or not isinstance(name, str):
        msg = f"invalid provider file {source}: provider_id, provider_code, display_name required"
        raise ValueError(msg)

    endpoint = raw.get("api_endpoint")
    policies = raw.get("supported_policies") or []
    if not isinstance(policies, list):
        msg = f"supported_policies must be a list in {source}"
        raise ValueError(msg)
    status = str(raw.get("service_status") or "active")
    extra = raw.get("extra") or {}
    extra_json = json.dumps(extra, sort_keys=True) if isinstance(extra, dict) else "{}"
    cref_raw = raw.get("credentials_ref")
    cref = str(cref_raw).strip() if cref_raw else None

    coverage = raw.get("coverage") or []
    carriers_out: list[ConnectableCarrier] = []
    if not isinstance(coverage, list):
        msg = f"coverage must be a list in {source}"
        raise ValueError(msg)
    ic = implicit_country.strip().upper() if implicit_country else None
    for block in coverage:
        if not isinstance(block, dict):
            continue
        explicit = block.get("country")
        if explicit is not None and explicit != "":
            cc = str(explicit).strip().upper()
            if ic is not None and cc != ic:
                msg = f"{source}: coverage.country {cc} does not match partition {ic}"
                raise ValueError(msg)
        elif ic is not None:
            cc = ic
        else:
            msg = f"{source}: set coverage.country or place file under countries/<ISO>/"
            raise ValueError(msg)
        carrs = block.get("carriers") or []
        if not isinstance(carrs, list):
            continue
        for c in carrs:
            carriers_out.append(
                ConnectableCarrier(country_code=cc, carrier_code=str(c).strip())
            )

    return ProviderRegistryConfig(
        provider_id=pid,
        provider_code=pcode,
        display_name=name,
        api_endpoint=str(endpoint) if endpoint else None,
        supported_policies=[str(p) for p in policies],
        service_status=status,
        extra_config_json=extra_json,
        connectable_carriers=carriers_out,
        credentials_ref=cref,
    )


def _merge_same_provider(
    left: ProviderRegistryConfig,
    right: ProviderRegistryConfig,
    *,
    right_source: Path,
) -> ProviderRegistryConfig:
    if left.provider_id != right.provider_id:
        msg = "internal merge id mismatch"
        raise ValueError(msg)
    if left.extra_config_json != right.extra_config_json or left.api_endpoint != right.api_endpoint:
        logger.warning(
            "provider_registry_metadata_drift",
            extra={
                "provider_id": left.provider_id,
                "fragment": str(right_source),
                "note": "using first fragment metadata; align YAML copies per country",
            },
        )
    seen: set[tuple[str, str]] = set()
    merged: list[ConnectableCarrier] = []
    for c in (*left.connectable_carriers, *right.connectable_carriers):
        key = (c.country_code, c.carrier_code)
        if key in seen:
            continue
        seen.add(key)
        merged.append(c)
    merged_cref = left.credentials_ref or right.credentials_ref
    if (
        left.credentials_ref
        and right.credentials_ref
        and left.credentials_ref != right.credentials_ref
    ):
        logger.warning(
            "provider_registry_credentials_ref_mismatch",
            extra={
                "provider_id": left.provider_id,
                "left": left.credentials_ref,
                "right": right.credentials_ref,
                "fragment": str(right_source),
            },
        )
    return ProviderRegistryConfig(
        provider_id=left.provider_id,
        provider_code=left.provider_code,
        display_name=left.display_name,
        api_endpoint=left.api_endpoint,
        supported_policies=list(left.supported_policies),
        service_status=left.service_status,
        extra_config_json=left.extra_config_json,
        connectable_carriers=merged,
        credentials_ref=merged_cref,
    )


def _load_partitioned_catalog(root: Path, index: dict[str, Any]) -> list[ProviderRegistryConfig]:
    catalogs = index.get("country_catalogs")
    if not isinstance(catalogs, list) or not catalogs:
        msg = f"country_catalogs must be a non-empty list in {root / _INDEX_NAME}"
        raise ValueError(msg)

    by_id: dict[str, ProviderRegistryConfig] = {}
    for rel in catalogs:
        if not isinstance(rel, str):
            continue
        cat_path = (root / rel).resolve()
        if not cat_path.is_file():
            msg = f"country catalog not found: {cat_path}"
            raise FileNotFoundError(msg)
        cidx = _read_yaml(cat_path)
        if not isinstance(cidx, dict):
            msg = f"invalid country index: {cat_path}"
            raise ValueError(msg)
        country_dir = cat_path.parent
        implicit_cc = country_dir.name.strip().upper()
        files = cidx.get("provider_files")
        if not isinstance(files, list):
            msg = f"provider_files missing in {cat_path}"
            raise ValueError(msg)
        for pf in files:
            if not isinstance(pf, str):
                continue
            path = (country_dir / pf).resolve()
            if not path.is_file():
                msg = f"provider fragment not found: {path}"
                raise FileNotFoundError(msg)
            raw = _read_yaml(path)
            if not isinstance(raw, dict):
                msg = f"expected mapping in {path}"
                raise ValueError(msg)
            cfg = _parse_provider_doc(raw, path, implicit_country=implicit_cc)
            existing = by_id.get(cfg.provider_id)
            if existing is None:
                by_id[cfg.provider_id] = cfg
            else:
                by_id[cfg.provider_id] = _merge_same_provider(
                    existing, cfg, right_source=path
                )

    return sorted(by_id.values(), key=lambda c: c.provider_id)


def _load_legacy_flat_catalog(root: Path, index: dict[str, Any]) -> list[ProviderRegistryConfig]:
    files = index.get("provider_files")
    if not isinstance(files, list) or not files:
        msg = f"provider_files must be a non-empty list in {root / _INDEX_NAME}"
        raise ValueError(msg)
    out: list[ProviderRegistryConfig] = []
    for rel in files:
        if not isinstance(rel, str):
            continue
        path = (root / rel).resolve()
        if not path.is_file():
            msg = f"provider registry fragment not found: {path}"
            raise FileNotFoundError(msg)
        raw = _read_yaml(path)
        if not isinstance(raw, dict):
            msg = f"expected mapping in {path}"
            raise ValueError(msg)
        out.append(_parse_provider_doc(raw, path, implicit_country=None))
    return out


@lru_cache(maxsize=1)
def load_all_provider_configs() -> tuple[ProviderRegistryConfig, ...]:
    """Load and parse the full catalog (cached; restart process to pick up edits)."""
    root = registry_root()
    index_path = root / _INDEX_NAME
    if not index_path.is_file():
        msg = f"provider registry index missing: {index_path}"
        raise FileNotFoundError(msg)

    index = _read_yaml(index_path)
    if not isinstance(index, dict):
        msg = f"invalid registry index: {index_path}"
        raise ValueError(msg)

    if index.get("country_catalogs"):
        rows = _load_partitioned_catalog(root, index)
    elif index.get("provider_files"):
        rows = _load_legacy_flat_catalog(root, index)
    else:
        msg = f"index must define country_catalogs (v2) or provider_files (legacy): {index_path}"
        raise ValueError(msg)

    return tuple(rows)


def clear_registry_cache() -> None:
    """Test helper: invalidate cached YAML."""
    load_all_provider_configs.cache_clear()
