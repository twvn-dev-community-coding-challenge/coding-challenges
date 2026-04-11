"""Tests for PII redaction utilities."""

from __future__ import annotations

from py_core.redact import (
    DEFAULT_POLICY,
    RedactionPolicy,
    mask_email,
    mask_phone,
    partial_mask,
    redact_full,
    redact_value,
)


def test_mask_phone_with_country_code() -> None:
    assert mask_phone("+84912345678") == "+84*****5678"


def test_mask_phone_short_number() -> None:
    assert mask_phone("+1234") == "[REDACTED]"


def test_mask_phone_without_plus() -> None:
    assert mask_phone("0912345678") == "091***5678"


def test_mask_email_normal() -> None:
    assert mask_email("nguyen@example.com") == "n***n@e***.com"


def test_mask_email_short_local() -> None:
    assert mask_email("a@b.com") == "a***@b***.com"


def test_mask_email_invalid() -> None:
    assert mask_email("not-an-email") == "[REDACTED]"


def test_redact_full() -> None:
    assert redact_full("x") == "[REDACTED]"
    assert redact_full(42) == "[REDACTED]"


def test_partial_mask() -> None:
    assert partial_mask("nguyen") == "n***n"


def test_partial_mask_short() -> None:
    assert partial_mask("ab") == "[REDACTED]"


def test_redact_dict_simple() -> None:
    policy = DEFAULT_POLICY
    data = {"phone_number": "+84912345678", "name": "John"}
    out = policy.redact_dict(data)
    assert out == {"phone_number": "+84*****5678", "name": "John"}


def test_redact_dict_nested() -> None:
    policy = DEFAULT_POLICY
    data = {"channel_payload": {"phone_number": "+84912345678"}}
    out = policy.redact_dict(data)
    assert out == {"channel_payload": {"phone_number": "+84*****5678"}}


def test_redact_dict_list() -> None:
    policy = DEFAULT_POLICY
    data = {"items": [{"phone_number": "+841234567890"}]}
    out = policy.redact_dict(data)
    assert out == {"items": [{"phone_number": "+84******7890"}]}


def test_redact_dict_content_field_kept() -> None:
    policy = DEFAULT_POLICY
    data = {"content": "Hello world"}
    out = policy.redact_dict(data)
    assert out == {"content": "Hello world"}


def test_redact_dict_password_regex() -> None:
    policy = DEFAULT_POLICY
    data = {"password": "secret123"}
    out = policy.redact_dict(data)
    assert out == {"password": "[REDACTED]"}


def test_redact_dict_api_key_regex() -> None:
    policy = DEFAULT_POLICY
    data = {"api_key": "sk-abc123"}
    out = policy.redact_dict(data)
    assert out == {"api_key": "[REDACTED]"}


def test_redact_dict_no_mutation() -> None:
    policy = DEFAULT_POLICY
    data = {"phone_number": "+84912345678", "nested": {"x": 1}}
    snapshot = {"phone_number": "+84912345678", "nested": {"x": 1}}
    policy.redact_dict(data)
    assert data == snapshot


def test_default_policy_exists() -> None:
    assert isinstance(DEFAULT_POLICY, RedactionPolicy)
    assert len(DEFAULT_POLICY.rules) > 0


def test_redact_value_public_api() -> None:
    assert redact_value("+84912345678", "mask_phone") == "+84*****5678"
