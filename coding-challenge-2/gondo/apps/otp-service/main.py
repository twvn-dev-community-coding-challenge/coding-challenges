"""OTP service HTTP API — server-side codes with TTL and verify."""

from __future__ import annotations

import os
import uuid

from fastapi import HTTPException, status
from pydantic import BaseModel, Field

from py_core.app import create_app
from py_core.logging import _should_log_payloads, configure_logging

from service import VerifyOutcome, issue_challenge, verify_challenge

_LOGS_DIR = os.path.join(os.path.dirname(__file__), "..", "..", "logs")
configure_logging(service_name="otp-service", log_dir=_LOGS_DIR)

_PAYLOAD_LOGGING = _should_log_payloads(None)


class IssueChallengeRequest(BaseModel):
    subject: str = Field(..., min_length=1, max_length=512, description="Correlation id e.g. message_id")


class IssueChallengeResponse(BaseModel):
    challenge_id: str
    expires_at: str
    ttl_seconds: int
    code: str = Field(
        ...,
        description="Plaintext OTP for SMS composition; do not expose to browsers in production.",
    )


class VerifyRequest(BaseModel):
    challenge_id: str
    code: str = Field(..., min_length=4, max_length=32)


class VerifyResponse(BaseModel):
    status: str


app = create_app(
    title="OTP Service",
    description="Issue and verify time-bound OTP codes (hashed at rest, TTL, attempt limits).",
    service_name="otp-service",
    enable_payload_logging=_PAYLOAD_LOGGING,
)


@app.post("/v1/challenges", response_model=IssueChallengeResponse)
async def issue(body: IssueChallengeRequest) -> IssueChallengeResponse:
    cid, code, expires_at, ttl_seconds = await issue_challenge(subject=body.subject)
    return IssueChallengeResponse(
        challenge_id=str(cid),
        expires_at=expires_at.isoformat(),
        ttl_seconds=ttl_seconds,
        code=code,
    )


@app.post("/v1/verify", response_model=VerifyResponse)
async def verify(body: VerifyRequest) -> VerifyResponse:
    try:
        cid = uuid.UUID(body.challenge_id.strip())
    except ValueError:
        raise HTTPException(
            status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail="invalid_challenge_id",
        )
    outcome = await verify_challenge(challenge_id=cid, code=body.code)
    match outcome:
        case VerifyOutcome.SUCCESS:
            return VerifyResponse(status="verified")
        case VerifyOutcome.NOT_FOUND:
            raise HTTPException(status.HTTP_404_NOT_FOUND, detail="challenge_not_found")
        case VerifyOutcome.EXPIRED:
            raise HTTPException(status.HTTP_410_GONE, detail="challenge_expired")
        case VerifyOutcome.EXHAUSTED:
            raise HTTPException(status.HTTP_423_LOCKED, detail="attempts_exhausted")
        case VerifyOutcome.INVALID_CODE:
            raise HTTPException(status.HTTP_401_UNAUTHORIZED, detail="invalid_code")
        case VerifyOutcome.ALREADY_CONSUMED:
            raise HTTPException(
                status.HTTP_409_CONFLICT,
                detail="challenge_already_consumed",
            )
