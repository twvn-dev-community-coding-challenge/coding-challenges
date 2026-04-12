"""PostgreSQL access for notification-service.

The ``notification_service`` schema has no application-owned tables after the TD
refactor; carrier resolution reads from ``provider_service.carrier_prefixes``
joined with ``provider_service.carriers``.
"""

from __future__ import annotations

import sqlalchemy as sa
from sqlalchemy.ext.asyncio import AsyncSession

_CARRIER_PREFIXES_TABLE = sa.table(
    "carrier_prefixes",
    sa.column("id", sa.Uuid),
    sa.column("country_calling_code", sa.String),
    sa.column("national_destination", sa.String),
    sa.column("carrier_id", sa.Integer),
    sa.column("match_priority", sa.Integer),
    schema="provider_service",
)

_CARRIERS_TABLE = sa.table(
    "carriers",
    sa.column("id", sa.Integer),
    sa.column("code", sa.String),
    schema="provider_service",
)


async def resolve_carrier(session: AsyncSession, phone_number: str) -> str:
    """Resolve carrier code from phone number via ``provider_service`` tables.

    Joins ``provider_service.carrier_prefixes`` with ``provider_service.carriers``
    to return the carrier ``code`` (e.g. ``VIETTEL``).

    Algorithm:
    1. Strip '+' from phone_number
    2. Load prefix rows ordered by match_priority DESC, national_destination length DESC
    3. For each row, if digits start with country_calling_code + national_destination,
       return that row's carrier code (first match wins — equivalent to highest
       priority then longest prefix)
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
            _CARRIER_PREFIXES_TABLE.c.match_priority,
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
