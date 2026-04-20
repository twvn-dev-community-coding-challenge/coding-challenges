"""Tests for phone normalization and validation."""

from __future__ import annotations

from fastapi.testclient import TestClient

from main import app
from phone_validation import normalize_phone_number, validate_phone_number


def test_normalize_vn_national() -> None:
    assert normalize_phone_number("VN", "0901234567") == "+84901234567"


def test_normalize_vn_e164_unchanged_shape() -> None:
    assert normalize_phone_number("VN", "+84901234567") == "+84901234567"


def test_normalize_vn_strips_formatting() -> None:
    assert normalize_phone_number("VN", "(090) 123-4567") == "+84901234567"


def test_normalize_th_national() -> None:
    assert normalize_phone_number("TH", "0812345678") == "+66812345678"


def test_normalize_th_e164() -> None:
    assert normalize_phone_number("TH", "+66812345678") == "+66812345678"


def test_normalize_sg_national_eight_digits() -> None:
    assert normalize_phone_number("SG", "91234567") == "+6591234567"


def test_normalize_sg_e164() -> None:
    assert normalize_phone_number("SG", "+6591234567") == "+6591234567"


def test_normalize_ph_national() -> None:
    assert normalize_phone_number("PH", "09171234567") == "+639171234567"


def test_normalize_ph_e164() -> None:
    assert normalize_phone_number("PH", "+639171234567") == "+639171234567"


def test_validate_vn_mobifone_ok() -> None:
    assert validate_phone_number("VN", "+84901234567") is None


def test_validate_vn_viettel_mock_scenario_ok() -> None:
    assert validate_phone_number("VN", "+84399000001") is None


def test_validate_th_ais_ok() -> None:
    assert validate_phone_number("TH", "+66812345678") is None


def test_validate_sg_starhub_ok() -> None:
    assert validate_phone_number("SG", "+6591234567") is None


def test_validate_ph_globe_ok() -> None:
    assert validate_phone_number("PH", "+639171234567") is None


def test_validate_vn_invalid_prefix() -> None:
    # Nine digits after +84 so length passes; "12" is not a valid VN carrier pair.
    err = validate_phone_number("VN", "+84123456789")
    assert err is not None
    assert "12" in err or "prefix" in err.lower()


def test_validate_vn_too_short() -> None:
    err = validate_phone_number("VN", "+8490123")
    assert err is not None
    assert "9" in err


def test_validate_vn_too_long() -> None:
    err = validate_phone_number("VN", "+849012345678")
    assert err is not None


def test_validate_th_invalid_prefix() -> None:
    err = validate_phone_number("TH", "+66912345678")
    assert err is not None
    assert "81" in err or "82" in err


def test_validate_sg_invalid_first_digit() -> None:
    err = validate_phone_number("SG", "+6571234567")
    assert err is not None
    assert "8" in err and "9" in err


def test_validate_ph_invalid_prefix() -> None:
    err = validate_phone_number("PH", "+639001234567")
    assert err is not None
    assert "905" in err or "prefix" in err.lower()


def test_validate_unknown_country_skipped() -> None:
    assert validate_phone_number("XX", "+15551234567") is None
    assert validate_phone_number("XX", "garbage") is None


def test_integration_post_vn_national_stores_e164() -> None:
    client = TestClient(app)
    response = client.post(
        "/notifications",
        json={
            "message_id": "msg-phone-vn-national",
            "channel_type": "SMS",
            "recipient": "u1",
            "content": "hi",
            "channel_payload": {"country_code": "VN", "phone_number": "0901234567"},
        },
    )
    assert response.status_code == 201
    assert response.json()["data"]["channel_payload"]["phone_number"] == "+84901234567"


def test_integration_post_vn_nine_digits_rejected() -> None:
    client = TestClient(app)
    response = client.post(
        "/notifications",
        json={
            "message_id": "msg-phone-vn-9",
            "channel_type": "SMS",
            "recipient": "u2",
            "content": "hi",
            "channel_payload": {"country_code": "VN", "phone_number": "999999999"},
        },
    )
    assert response.status_code == 422
    assert response.json()["error"]["code"] == "VALIDATION_ERROR"


def test_integration_post_invalid_vn_returns_422() -> None:
    client = TestClient(app)
    response = client.post(
        "/notifications",
        json={
            "message_id": "msg-phone-vn-bad",
            "channel_type": "SMS",
            "recipient": "u3",
            "content": "hi",
            # Too few digits after +84 (invalid structure for VN).
            "channel_payload": {"country_code": "VN", "phone_number": "+8412345678"},
        },
    )
    assert response.status_code == 422
    body = response.json()
    assert body["error"]["code"] == "VALIDATION_ERROR"
    assert "phone_number" in body["error"]["details"]
