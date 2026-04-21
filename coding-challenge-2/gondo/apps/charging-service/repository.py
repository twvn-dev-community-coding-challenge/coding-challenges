"""PostgreSQL repository for charging-service (charging_service schema)."""

from __future__ import annotations

from datetime import datetime

import sqlalchemy as sa
from sqlalchemy.ext.asyncio import AsyncSession

_RATES_TABLE = sa.table(
    "rates",
    sa.column("id", sa.Uuid(as_uuid=True)),
    sa.column("country_code", sa.String),
    sa.column("carrier_id", sa.Integer),
    sa.column("provider_id", sa.Integer),
    sa.column("currency", sa.String),
    sa.column("price_per_sms", sa.Numeric(12, 6)),
    sa.column("effective_from", sa.DateTime(timezone=True)),
    sa.column("effective_to", sa.DateTime(timezone=True)),
    schema="charging_service",
)

_PROVIDERS_TABLE = sa.table(
    "providers",
    sa.column("id", sa.Integer),
    sa.column("code", sa.String),
    schema="provider_service",
)

_CARRIERS_TABLE = sa.table(
    "carriers",
    sa.column("id", sa.Integer),
    sa.column("code", sa.String),
    schema="provider_service",
)


async def fetch_provider_id_by_code(session: AsyncSession, code: str) -> int | None:
    """Resolve provider business code (e.g. prv_01) to integer id."""
    stmt = sa.select(_PROVIDERS_TABLE.c.id).where(_PROVIDERS_TABLE.c.code == code).limit(1)
    result = await session.execute(stmt)
    row = result.first()
    return int(row[0]) if row is not None else None


async def fetch_carrier_id_by_code(session: AsyncSession, code: str) -> int | None:
    """Resolve carrier code (e.g. VIETTEL) to integer id."""
    stmt = sa.select(_CARRIERS_TABLE.c.id).where(_CARRIERS_TABLE.c.code == code).limit(1)
    result = await session.execute(stmt)
    row = result.first()
    return int(row[0]) if row is not None else None


async def fetch_active_rate(
    session: AsyncSession,
    provider_id: int,
    country_code: str,
    carrier_id: int,
    as_of: datetime,
) -> sa.Row | None:
    """Find the active rate for provider/country/carrier at as_of."""
    stmt = (
        sa.select(_RATES_TABLE)
        .where(
            sa.and_(
                _RATES_TABLE.c.provider_id == provider_id,
                _RATES_TABLE.c.country_code == country_code,
                _RATES_TABLE.c.carrier_id == carrier_id,
                _RATES_TABLE.c.effective_from <= as_of,
                sa.or_(
                    _RATES_TABLE.c.effective_to.is_(None),
                    _RATES_TABLE.c.effective_to > as_of,
                ),
            )
        )
        .order_by(_RATES_TABLE.c.effective_from.desc())
        .limit(1)
    )
    result = await session.execute(stmt)
    return result.first()
