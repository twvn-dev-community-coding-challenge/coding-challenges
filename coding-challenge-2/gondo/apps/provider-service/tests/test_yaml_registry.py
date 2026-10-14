"""Tests for infra YAML provider registry loading."""

from __future__ import annotations

from pathlib import Path

import pytest

import yaml_registry


def test_load_default_registry_includes_twilio_vn(monkeypatch: pytest.MonkeyPatch) -> None:
    yaml_registry.load_all_provider_configs.cache_clear()
    root = Path(__file__).resolve().parents[3] / "infra" / "provider-registry"
    monkeypatch.setenv("PROVIDER_REGISTRY_ROOT", str(root))
    yaml_registry.load_all_provider_configs.cache_clear()
    configs = yaml_registry.load_all_provider_configs()
    twilio = next(c for c in configs if c.provider_id == "prv_01")
    assert twilio.api_endpoint == "https://api.twilio.com"
    countries = {x.country_code for x in twilio.connectable_carriers}
    assert "VN" in countries
    assert "SG" in countries
    assert len(twilio.connectable_carriers) >= 4
    monkeypatch.delenv("PROVIDER_REGISTRY_ROOT", raising=False)
    yaml_registry.load_all_provider_configs.cache_clear()


def test_parse_minimal_provider_file(tmp_path: Path, monkeypatch: pytest.MonkeyPatch) -> None:
    reg = tmp_path / "registry"
    (reg / "providers").mkdir(parents=True)
    (reg / "index.yaml").write_text(
        "schema_version: '1'\nkind: x\nprovider_files:\n  - providers/p.yaml\n",
        encoding="utf-8",
    )
    (reg / "providers" / "p.yaml").write_text(
        """
provider_id: prv_x
provider_code: TEST
display_name: Test Provider
api_endpoint: https://example.test/sms
supported_policies: [highest_precedence]
service_status: active
extra: { "k": "v" }
coverage:
  - country: VN
    carriers: [VIETTEL]
""",
        encoding="utf-8",
    )
    monkeypatch.setenv("PROVIDER_REGISTRY_ROOT", str(reg))
    yaml_registry.load_all_provider_configs.cache_clear()
    cfg = yaml_registry.load_all_provider_configs()[0]
    assert cfg.provider_id == "prv_x"
    assert '"k"' in cfg.extra_config_json
    assert cfg.connectable_carriers[0].carrier_code == "VIETTEL"
    monkeypatch.delenv("PROVIDER_REGISTRY_ROOT", raising=False)
    yaml_registry.load_all_provider_configs.cache_clear()


def test_partitioned_infobip_merges_vn_and_th(monkeypatch: pytest.MonkeyPatch) -> None:
    yaml_registry.load_all_provider_configs.cache_clear()
    root = Path(__file__).resolve().parents[3] / "infra" / "provider-registry"
    monkeypatch.setenv("PROVIDER_REGISTRY_ROOT", str(root))
    yaml_registry.load_all_provider_configs.cache_clear()
    infobip = next(c for c in yaml_registry.load_all_provider_configs() if c.provider_id == "prv_03")
    codes = {(x.country_code, x.carrier_code) for x in infobip.connectable_carriers}
    assert ("VN", "MOBIFONE") in codes
    assert ("TH", "AIS") in codes
    monkeypatch.delenv("PROVIDER_REGISTRY_ROOT", raising=False)
    yaml_registry.load_all_provider_configs.cache_clear()
