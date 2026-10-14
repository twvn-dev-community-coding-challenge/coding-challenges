"""Tests for ``cqrs.mock_scenario_loader``."""

from __future__ import annotations

import json
from pathlib import Path

import pytest

from cqrs import mock_scenario_loader as msl


def test_scenarios_index_built_at_import() -> None:
    assert len(msl._MOCK_SCENARIO_BY_DIGITS) > 0
    assert len(msl._SCENARIOS) == len(msl._MOCK_SCENARIO_BY_DIGITS)


def test_lookup_known_vn_scenario() -> None:
    s = msl.lookup_mock_scenario("+84399000001", "VN")
    assert s is not None
    assert s.get("outcome") == "send_success"
    assert s.get("country_code") == "VN"


def test_lookup_legacy_dev_defaults() -> None:
    s = msl.lookup_mock_scenario("+84999999999", "VN")
    assert s is not None
    assert s.get("outcome") == "send_success"
    s2 = msl.lookup_mock_scenario("+8490909090", "VN")
    assert s2 is not None
    assert s2.get("outcome") == "send_success"


def test_lookup_unknown_phone_returns_none() -> None:
    assert msl.lookup_mock_scenario("+00000000000", "US") is None
    assert msl.lookup_mock_scenario("+84123456789", "VN") is None


@pytest.mark.parametrize(
    ("phone", "country", "expected"),
    [
        ("+84399000001", "VN", "send_success"),
        ("+84395000005", "VN", "carrier_rejected"),
        ("+84396000004", "VN", "send_failed"),
        ("+84394000006", "VN", "retry_then_success"),
        ("+66823000044", "TH", "carrier_rejected"),
        ("+639182000083", "PH", "carrier_rejected"),
    ],
)
def test_get_mock_outcome_for_each_type(phone: str, country: str, expected: str) -> None:
    assert msl.get_mock_outcome(phone, country) == expected


def test_get_mock_actual_cost_present_and_null() -> None:
    assert msl.get_mock_actual_cost("+84399000001", "VN") == pytest.approx(0.0175)
    assert msl.get_mock_actual_cost("+84396000004", "VN") is None


def test_json_file_path_exists_next_to_package() -> None:
    expected = Path(__file__).resolve().parents[1] / "mock_scenarios.json"
    assert msl._MOCK_SCENARIOS_PATH == expected
    assert expected.is_file()
    with expected.open(encoding="utf-8") as f:
        root = json.load(f)
    assert isinstance(root.get("scenarios"), list)
