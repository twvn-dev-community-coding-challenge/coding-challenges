"""Carrier-registry HTTP API (bounded context separate from provider-service)."""

from __future__ import annotations

import json
from pathlib import Path

import pytest
from fastapi.testclient import TestClient

from py_core.credentials_store import clear_credentials_cache

from main import app
from registry_loader import clear_carrier_registry_cache


@pytest.fixture(autouse=True)
def _clear_registry_cache() -> None:
    clear_credentials_cache()
    clear_carrier_registry_cache()
    yield
    clear_carrier_registry_cache()
    clear_credentials_cache()


def test_registry_carriers_vn_viettel(
    monkeypatch: pytest.MonkeyPatch,
    tmp_path: Path,
) -> None:
    root = tmp_path / "carrier-registry"
    (root / "countries" / "VN").mkdir(parents=True)
    (root / "index.yaml").write_text(
        "schema_version: '1'\nkind: CarrierRegistryCatalog\ncountry_catalogs:\n  - countries/VN/index.yaml\n",
        encoding="utf-8",
    )
    (root / "countries" / "VN" / "index.yaml").write_text(
        "carrier_files:\n  - VIETTEL.yaml\n",
        encoding="utf-8",
    )
    (root / "countries" / "VN" / "VIETTEL.yaml").write_text(
        """country_code: VN
carrier_code: VIETTEL
carrier_credentials_ref: carrier/VN/VIETTEL
routing_hints:
  max_concat_segments: \"6\"
""",
        encoding="utf-8",
    )
    secrets = tmp_path / "secrets.json"
    secrets.write_text(
        json.dumps({"carrier/VN/VIETTEL": {"mno_webhook_secret": "x"}}),
        encoding="utf-8",
    )
    monkeypatch.setenv("CARRIER_REGISTRY_ROOT", str(root))
    monkeypatch.setenv("CREDENTIALS_BACKEND", "mock")
    monkeypatch.setenv("MOCK_CREDENTIALS_PATH", str(secrets))
    clear_carrier_registry_cache()

    client = TestClient(app)
    resp = client.get("/registry/carriers", params={"country_code": "VN"})
    assert resp.status_code == 200
    body = resp.json()
    assert body["country_code"] == "VN"
    assert body["credentials_backend"] == "mock"
    entries = body["entries"]
    assert len(entries) == 1
    assert entries[0]["carrier_code"] == "VIETTEL"
    assert entries[0]["routing_hints"]["max_concat_segments"] == "6"
    assert entries[0]["carrier_credentials_ref"] == "carrier/VN/VIETTEL"
    assert entries[0]["carrier_credentials_configured"] is True


def test_registry_carriers_filter_by_carrier(
    monkeypatch: pytest.MonkeyPatch,
    tmp_path: Path,
) -> None:
    root = tmp_path / "carrier-registry"
    (root / "countries" / "VN").mkdir(parents=True)
    (root / "index.yaml").write_text(
        "schema_version: '1'\nkind: x\ncountry_catalogs:\n  - countries/VN/index.yaml\n",
        encoding="utf-8",
    )
    (root / "countries" / "VN" / "index.yaml").write_text(
        "carrier_files:\n  - A.yaml\n  - B.yaml\n",
        encoding="utf-8",
    )
    for code in ("A", "B"):
        (root / "countries" / "VN" / f"{code}.yaml").write_text(
            f"country_code: VN\ncarrier_code: {code}\nrouting_hints: {{}}\n",
            encoding="utf-8",
        )
    monkeypatch.setenv("CARRIER_REGISTRY_ROOT", str(root))
    monkeypatch.setenv("CREDENTIALS_BACKEND", "mock")
    monkeypatch.delenv("MOCK_CREDENTIALS_PATH", raising=False)
    clear_carrier_registry_cache()

    client = TestClient(app)
    resp = client.get(
        "/registry/carriers",
        params={"country_code": "VN", "carrier": "B"},
    )
    assert resp.status_code == 200
    codes = {e["carrier_code"] for e in resp.json()["entries"]}
    assert codes == {"B"}
