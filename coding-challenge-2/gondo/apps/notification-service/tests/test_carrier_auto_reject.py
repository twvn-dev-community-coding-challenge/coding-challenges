"""Unit tests for simulated Carrier-rejected MSISDN."""

from __future__ import annotations

from cqrs.carrier_auto_reject import should_auto_carrier_reject, vn_msisdn_digits


def test_vn_msisdn_digits_national() -> None:
    assert vn_msisdn_digits("090909094") == "8490909094"


def test_vn_msisdn_digits_e164() -> None:
    assert vn_msisdn_digits("+8490909094") == "8490909094"


def test_should_auto_reject_vn_magic_number() -> None:
    assert should_auto_carrier_reject("VN", "090909094") is True
    assert should_auto_carrier_reject("VN", "+8490909094") is True


def test_should_not_reject_other_vn_numbers() -> None:
    assert should_auto_carrier_reject("VN", "090909095") is False
    assert should_auto_carrier_reject("VN", "+84901234567") is False


def test_should_not_reject_non_vn() -> None:
    assert should_auto_carrier_reject("US", "090909094") is False
