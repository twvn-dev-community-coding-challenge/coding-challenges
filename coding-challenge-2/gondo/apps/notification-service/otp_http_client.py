"""HTTP client to otp-service for server-side OTP issuance."""

from __future__ import annotations

import logging
import os
from typing import Any

import httpx

logger = logging.getLogger(__name__)

OTP_SERVICE_BASE_URL = os.environ.get(
    "OTP_SERVICE_BASE_URL",
    "http://127.0.0.1:8007",
).rstrip("/")


class OtpIssueError(Exception):
    """Raised when otp-service cannot issue a challenge."""


async def issue_challenge_http(subject: str) -> dict[str, Any]:
    """POST /v1/challenges; returns JSON body with challenge_id, code, expires_at, ttl_seconds."""
    url = f"{OTP_SERVICE_BASE_URL}/v1/challenges"
    try:
        async with httpx.AsyncClient(timeout=15.0) as client:
            r = await client.post(url, json={"subject": subject})
            r.raise_for_status()
    except httpx.HTTPStatusError as exc:
        logger.exception(
            "otp_issue_http_error",
            extra={"status": exc.response.status_code, "body": exc.response.text[:500]},
        )
        raise OtpIssueError(f"otp_service_http_{exc.response.status_code}") from exc
    except httpx.RequestError as exc:
        logger.exception("otp_issue_transport_error", extra={"detail": str(exc)})
        raise OtpIssueError("otp_service_unreachable") from exc
    data = r.json()
    for key in ("challenge_id", "code", "expires_at", "ttl_seconds"):
        if key not in data:
            raise OtpIssueError(f"otp_response_missing_{key}")
    return data


def substitute_otp_in_content(content: str, otp: str) -> str:
    if "{{OTP}}" in content:
        return content.replace("{{OTP}}", otp)
    return f"{content} OTP: {otp}"
