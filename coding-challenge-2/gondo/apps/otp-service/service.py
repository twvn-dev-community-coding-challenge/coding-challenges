"""Issue and verify OTP challenges."""

from __future__ import annotations

import uuid
from datetime import datetime, timedelta, timezone
from enum import Enum

from py_core.db import get_session

from otp_crypto import generate_six_digit_code, hash_code, verify_hash
from repository import decrement_attempts, insert_challenge, lock_challenge_row, mark_consumed
from settings import otp_hash_pepper, otp_max_attempts, otp_ttl_seconds


class VerifyOutcome(str, Enum):
    SUCCESS = "SUCCESS"
    NOT_FOUND = "NOT_FOUND"
    EXPIRED = "EXPIRED"
    EXHAUSTED = "EXHAUSTED"
    INVALID_CODE = "INVALID_CODE"
    ALREADY_CONSUMED = "ALREADY_CONSUMED"


async def issue_challenge(*, subject: str) -> tuple[uuid.UUID, str, datetime, int]:
    ttl = otp_ttl_seconds()
    attempts = otp_max_attempts()
    pepper = otp_hash_pepper()
    code = generate_six_digit_code()
    hashed = hash_code(pepper, code)
    cid = uuid.uuid4()
    expires_at = datetime.now(timezone.utc) + timedelta(seconds=ttl)
    async with get_session() as session:
        await insert_challenge(
            session,
            challenge_id=cid,
            subject=subject.strip(),
            code_hash=hashed,
            expires_at=expires_at,
            attempts_remaining=attempts,
        )
        await session.commit()
    return cid, code, expires_at, ttl


async def verify_challenge(*, challenge_id: uuid.UUID, code: str) -> VerifyOutcome:
    pepper = otp_hash_pepper()
    async with get_session() as session:
        async with session.begin():
            row = await lock_challenge_row(session, challenge_id)
            if row is None:
                return VerifyOutcome.NOT_FOUND
            if row.consumed_at is not None:
                return VerifyOutcome.ALREADY_CONSUMED
            now = datetime.now(timezone.utc)
            if now >= row.expires_at:
                return VerifyOutcome.EXPIRED
            if row.attempts_remaining <= 0:
                return VerifyOutcome.EXHAUSTED
            if verify_hash(pepper, row.code_hash, code):
                await mark_consumed(session, challenge_id)
                return VerifyOutcome.SUCCESS
            await decrement_attempts(session, challenge_id)
            return VerifyOutcome.INVALID_CODE
