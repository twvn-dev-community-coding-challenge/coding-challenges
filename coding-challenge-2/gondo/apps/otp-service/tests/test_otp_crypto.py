"""Unit tests for OTP hashing."""

from __future__ import annotations

from otp_crypto import hash_code, verify_hash


def test_verify_hash_matches_roundtrip() -> None:
    pepper = "test-pepper"
    code = "042681"
    h = hash_code(pepper, code)
    assert verify_hash(pepper, h, code)
    assert not verify_hash(pepper, h, "999999")
