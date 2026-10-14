"""Generate numeric OTPs and hash with application pepper (never store plaintext)."""

from __future__ import annotations

import hashlib
import hmac
import secrets


def generate_six_digit_code() -> str:
    return f"{secrets.randbelow(1_000_000):06d}"


def hash_code(pepper: str, code: str) -> str:
    payload = f"{pepper}\x1e{code.strip()}".encode("utf-8")
    return hashlib.sha256(payload).hexdigest()


def verify_hash(pepper: str, stored_hash: str, candidate: str) -> bool:
    computed = hash_code(pepper, candidate)
    return hmac.compare_digest(stored_hash, computed)
