"""Unit tests for simulated Carrier-rejected MSISDNs (mock_scenarios.json)."""

from __future__ import annotations

from cqrs.carrier_auto_reject import should_auto_carrier_reject
from cqrs.mock_scenario_loader import canonical_mock_msisdn_digits


def test_canonical_vn_national_carrier_reject_example() -> None:
    assert canonical_mock_msisdn_digits("VN", "0395000005") == "84395000005"


def test_canonical_vn_e164_carrier_reject_example() -> None:
    assert canonical_mock_msisdn_digits("VN", "+84395000005") == "84395000005"


def test_should_auto_reject_vn_scenario() -> None:
    assert should_auto_carrier_reject("VN", "0395000005") is True
    assert should_auto_carrier_reject("VN", "+84395000005") is True


def test_should_not_reject_send_success_vn() -> None:
    assert should_auto_carrier_reject("VN", "+84399000001") is False


def test_should_auto_reject_ph_smart_scenario() -> None:
    assert should_auto_carrier_reject("PH", "+639182000083") is True


def test_should_not_reject_non_scenario_number() -> None:
    assert should_auto_carrier_reject("VN", "+84901234567") is False
    assert should_auto_carrier_reject("US", "+15551234567") is False
