"""PostgreSQL persistence for OTP challenges (``otp_service`` schema)."""

from __future__ import annotations

import uuid
from dataclasses import dataclass
from datetime import datetime, timezone

import sqlalchemy as sa
from sqlalchemy.ext.asyncio import AsyncSession

SCHEMA = "otp_service"

_challenges = sa.table(
    "challenges",
    sa.column("id", sa.Uuid(as_uuid=True)),
    sa.column("subject", sa.String),
    sa.column("code_hash", sa.String),
    sa.column("expires_at", sa.DateTime(timezone=True)),
    sa.column("attempts_remaining", sa.Integer),
    sa.column("consumed_at", sa.DateTime(timezone=True)),
    sa.column("created_at", sa.DateTime(timezone=True)),
    schema=SCHEMA,
)


@dataclass(frozen=True)
class ChallengeRow:
    id: uuid.UUID
    subject: str
    code_hash: str
    expires_at: datetime
    attempts_remaining: int
    consumed_at: datetime | None


async def insert_challenge(
    session: AsyncSession,
    *,
    challenge_id: uuid.UUID,
    subject: str,
    code_hash: str,
    expires_at: datetime,
    attempts_remaining: int,
) -> None:
    now = datetime.now(timezone.utc)
    await session.execute(
        sa.insert(_challenges).values(
            id=challenge_id,
            subject=subject,
            code_hash=code_hash,
            expires_at=expires_at,
            attempts_remaining=attempts_remaining,
            consumed_at=None,
            created_at=now,
        )
    )


async def lock_challenge_row(
    session: AsyncSession, challenge_id: uuid.UUID
) -> ChallengeRow | None:
    stmt = (
        sa.select(_challenges)
        .where(_challenges.c.id == challenge_id)
        .with_for_update()
    )
    result = await session.execute(stmt)
    row = result.first()
    if row is None:
        return None
    r = row._mapping
    return ChallengeRow(
        id=r["id"],
        subject=r["subject"],
        code_hash=r["code_hash"],
        expires_at=r["expires_at"],
        attempts_remaining=int(r["attempts_remaining"]),
        consumed_at=r["consumed_at"],
    )


async def mark_consumed(session: AsyncSession, challenge_id: uuid.UUID) -> None:
    now = datetime.now(timezone.utc)
    await session.execute(
        sa.update(_challenges)
        .where(_challenges.c.id == challenge_id)
        .values(consumed_at=now)
    )


async def decrement_attempts(session: AsyncSession, challenge_id: uuid.UUID) -> None:
    await session.execute(
        sa.update(_challenges)
        .where(_challenges.c.id == challenge_id)
        .values(attempts_remaining=_challenges.c.attempts_remaining - 1)
    )
