"""Environment-backed OTP policy (TTL, attempts, pepper)."""

from __future__ import annotations

import os


def otp_ttl_seconds() -> int:
    raw = os.environ.get("OTP_TTL_SECONDS", "120")
    try:
        n = int(raw)
    except ValueError:
        return 120
    return max(30, min(n, 3600))


def otp_max_attempts() -> int:
    raw = os.environ.get("OTP_MAX_ATTEMPTS", "5")
    try:
        n = int(raw)
    except ValueError:
        return 5
    return max(1, min(n, 20))


def otp_hash_pepper() -> str:
    p = os.environ.get("OTP_HASH_PEPPER", "").strip()
    if p:
        return p
    return "dev-only-otp-pepper-change-me"


def otp_issue_requests_per_minute() -> int:
    """Max POST /v1/challenges per client IP per rolling minute; ``0`` disables."""
    raw = os.environ.get("OTP_ISSUE_REQUESTS_PER_MINUTE", "60")
    try:
        n = int(raw)
    except ValueError:
        return 60
    return max(0, min(n, 100_000))


def otp_verify_requests_per_minute() -> int:
    """Max POST /v1/verify per client IP per rolling minute; ``0`` disables."""
    raw = os.environ.get("OTP_VERIFY_REQUESTS_PER_MINUTE", "120")
    try:
        n = int(raw)
    except ValueError:
        return 120
    return max(0, min(n, 100_000))
