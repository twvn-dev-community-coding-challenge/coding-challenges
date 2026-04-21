"""Relax rate limits during the suite unless a test lowers them."""

from __future__ import annotations

import os

os.environ.setdefault("OTP_ISSUE_REQUESTS_PER_MINUTE", "50000")
os.environ.setdefault("OTP_VERIFY_REQUESTS_PER_MINUTE", "50000")
