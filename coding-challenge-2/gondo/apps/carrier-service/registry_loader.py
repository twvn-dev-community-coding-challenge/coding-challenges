"""Load ``infra/carrier-registry`` (country-partitioned carrier / MNO config)."""

from __future__ import annotations

import logging
import os
from functools import lru_cache
from pathlib import Path
from typing import Any

import yaml

from registry_types import CarrierRegistryEntry

_INDEX_NAME = "index.yaml"
logger = logging.getLogger(__name__)


def _default_carrier_registry_root() -> Path:
    base = Path(__file__).resolve().parent.parent.parent
    return base / "infra" / "carrier-registry"


def carrier_registry_root() -> Path:
    override = os.environ.get("CARRIER_REGISTRY_ROOT")
    if override:
        return Path(override).expanduser().resolve()
    return _default_carrier_registry_root()


def _read_yaml(path: Path) -> Any:
    return yaml.safe_load(path.read_text(encoding="utf-8"))


def _parse_carrier_file(path: Path, implicit_country: str) -> CarrierRegistryEntry:
    raw = _read_yaml(path)
    if not isinstance(raw, dict):
        msg = f"expected mapping in {path}"
        raise ValueError(msg)
    cc = str(raw.get("country_code") or implicit_country).strip().upper()
    carrier = str(raw.get("carrier_code", "")).strip()
    if not carrier:
        msg = f"carrier_code required in {path}"
        raise ValueError(msg)
    if cc != implicit_country:
        msg = f"{path}: country_code {cc} must match folder {implicit_country}"
        raise ValueError(msg)
    hints_raw = raw.get("routing_hints") or {}
    hints: dict[str, str] = {}
    if isinstance(hints_raw, dict):
        for k, v in hints_raw.items():
            hints[str(k)] = str(v)
    cref = raw.get("carrier_credentials_ref")
    cref_s = str(cref).strip() if cref else None
    return CarrierRegistryEntry(
        country_code=cc,
        carrier_code=carrier,
        routing_hints=hints,
        carrier_credentials_ref=cref_s,
    )


@lru_cache(maxsize=1)
def load_all_carrier_entries() -> tuple[CarrierRegistryEntry, ...]:
    root = carrier_registry_root()
    index_path = root / _INDEX_NAME
    if not index_path.is_file():
        logger.warning("carrier_registry_index_missing", extra={"path": str(index_path)})
        return tuple()
    index = _read_yaml(index_path)
    if not isinstance(index, dict):
        return tuple()
    catalogs = index.get("country_catalogs")
    if not isinstance(catalogs, list):
        return tuple()
    out: list[CarrierRegistryEntry] = []
    for rel in catalogs:
        if not isinstance(rel, str):
            continue
        cat_path = (root / rel).resolve()
        if not cat_path.is_file():
            continue
        cidx = _read_yaml(cat_path)
        if not isinstance(cidx, dict):
            continue
        country_dir = cat_path.parent
        implicit = country_dir.name.strip().upper()
        files = cidx.get("carrier_files")
        if not isinstance(files, list):
            continue
        for name in files:
            if not isinstance(name, str):
                continue
            path = (country_dir / name).resolve()
            if not path.is_file():
                msg = f"carrier registry fragment not found: {path}"
                raise FileNotFoundError(msg)
            out.append(_parse_carrier_file(path, implicit))
    return tuple(out)


def clear_carrier_registry_cache() -> None:
    load_all_carrier_entries.cache_clear()


def filter_carrier_entries(
    country_code: str,
    carrier: str | None,
) -> list[CarrierRegistryEntry]:
    cc = country_code.strip().upper()
    car = carrier.strip() if carrier else None
    rows = [e for e in load_all_carrier_entries() if e.country_code == cc]
    if car:
        rows = [e for e in rows if e.carrier_code == car]
    return rows
