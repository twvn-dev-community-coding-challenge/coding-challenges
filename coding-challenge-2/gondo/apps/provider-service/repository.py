"""PostgreSQL repository for provider-service (provider_service schema)."""

from __future__ import annotations

from datetime import datetime

import sqlalchemy as sa
from sqlalchemy.ext.asyncio import AsyncSession

_PROVIDERS_TABLE = sa.table(
    "providers",
    sa.column("id", sa.Integer),
    sa.column("code", sa.String),
    sa.column("name", sa.String),
    sa.column("created_at", sa.DateTime(timezone=True)),
    schema="provider_service",
)

_CARRIERS_TABLE = sa.table(
    "carriers",
    sa.column("id", sa.Integer),
    sa.column("code", sa.String),
    sa.column("display_name", sa.String),
    sa.column("country_code", sa.CHAR(2)),
    sa.column("status", sa.String),
    sa.column("created_at", sa.DateTime(timezone=True)),
    schema="provider_service",
)

_ROUTING_RULES_TABLE = sa.table(
    "routing_rules",
    sa.column("id", sa.Uuid(as_uuid=True)),
    sa.column("country_code", sa.CHAR(2)),
    sa.column("carrier_id", sa.Integer),
    sa.column("provider_id", sa.Integer),
    sa.column("priority", sa.Integer),
    sa.column("routing_rule_version", sa.Integer),
    sa.column("effective_from", sa.DateTime(timezone=True)),
    sa.column("effective_to", sa.DateTime(timezone=True)),
    schema="provider_service",
)

_CARRIER_PREFIXES_TABLE = sa.table(
    "carrier_prefixes",
    sa.column("id", sa.Uuid(as_uuid=True)),
    sa.column("country_calling_code", sa.String),
    sa.column("national_destination", sa.String),
    sa.column("carrier_id", sa.Integer),
    sa.column("match_priority", sa.Integer),
    schema="provider_service",
)


async def fetch_provider(session: AsyncSession, provider_id: int) -> sa.Row | None:
    result = await session.execute(
        sa.select(_PROVIDERS_TABLE).where(_PROVIDERS_TABLE.c.id == provider_id)
    )
    return result.first()


async def fetch_provider_by_code(session: AsyncSession, code: str) -> sa.Row | None:
    normalized = code.strip()
    result = await session.execute(
        sa.select(_PROVIDERS_TABLE).where(_PROVIDERS_TABLE.c.code == normalized)
    )
    return result.first()


async def fetch_all_providers(session: AsyncSession) -> list[sa.Row]:
    result = await session.execute(
        sa.select(_PROVIDERS_TABLE).order_by(_PROVIDERS_TABLE.c.id)
    )
    return list(result.all())


async def fetch_carrier_by_code(session: AsyncSession, code: str) -> sa.Row | None:
    normalized = code.strip()
    result = await session.execute(
        sa.select(_CARRIERS_TABLE).where(_CARRIERS_TABLE.c.code == normalized)
    )
    return result.first()


async def fetch_active_routing_rules(
    session: AsyncSession,
    country_code: str,
    carrier: str,
    as_of: datetime,
) -> list[sa.Row]:
    """Fetch routing rules matching country/carrier active at as_of, ordered by version DESC, priority DESC."""
    cc = country_code.strip().upper()
    carrier_code = carrier.strip()
    stmt = (
        sa.select(_ROUTING_RULES_TABLE)
        .select_from(
            _ROUTING_RULES_TABLE.join(
                _CARRIERS_TABLE,
                _ROUTING_RULES_TABLE.c.carrier_id == _CARRIERS_TABLE.c.id,
            )
        )
        .where(
            sa.and_(
                _ROUTING_RULES_TABLE.c.country_code == cc,
                _CARRIERS_TABLE.c.code == carrier_code,
                _ROUTING_RULES_TABLE.c.effective_from <= as_of,
                sa.or_(
                    _ROUTING_RULES_TABLE.c.effective_to.is_(None),
                    _ROUTING_RULES_TABLE.c.effective_to > as_of,
                ),
            )
        )
        .order_by(
            _ROUTING_RULES_TABLE.c.routing_rule_version.desc(),
            _ROUTING_RULES_TABLE.c.priority.desc(),
        )
    )
    result = await session.execute(stmt)
    return list(result.all())


async def resolve_carrier_from_prefix(session: AsyncSession, phone_number: str) -> str:
    """Resolve carrier code from phone number using carrier_prefixes joined with carriers.

    Algorithm:
    1. Strip '+' from phone_number
    2. Load prefix rows joined with carriers, ordered by match_priority DESC,
       national_destination length DESC
    3. For each row, if digits start with country_calling_code + national_destination,
       return that carrier's code (first match wins)
    4. Return 'UNKNOWN' if no match
    """
    digits = phone_number.lstrip("+")
    if not digits:
        return "UNKNOWN"

    stmt = (
        sa.select(
            _CARRIER_PREFIXES_TABLE.c.country_calling_code,
            _CARRIER_PREFIXES_TABLE.c.national_destination,
            _CARRIERS_TABLE.c.code.label("carrier_code"),
        )
        .select_from(
            _CARRIER_PREFIXES_TABLE.join(
                _CARRIERS_TABLE,
                _CARRIER_PREFIXES_TABLE.c.carrier_id == _CARRIERS_TABLE.c.id,
            )
        )
        .order_by(
            _CARRIER_PREFIXES_TABLE.c.match_priority.desc(),
            sa.func.length(_CARRIER_PREFIXES_TABLE.c.national_destination).desc(),
        )
    )
    result = await session.execute(stmt)
    rows = result.all()
    for row in rows:
        mapping = row._mapping
        cc = str(mapping["country_calling_code"])
        nd = str(mapping["national_destination"])
        full_prefix = cc + nd
        if digits.startswith(full_prefix):
            return str(mapping["carrier_code"])
    return "UNKNOWN"
