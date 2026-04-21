"""Rate table backed by PostgreSQL, estimates, and actual cost recording."""

from __future__ import annotations

import uuid
from dataclasses import dataclass
from datetime import datetime, timezone

import sqlalchemy as sa
from py_core.db import get_session
from py_core.proto_utils import as_utc
from repository import (
    fetch_active_rate,
    fetch_carrier_id_by_code,
    fetch_provider_id_by_code,
)


@dataclass(frozen=True)
class Rate:
    rate_id: str
    rate_version: int
    provider_id: str
    country_code: str
    carrier: str
    unit_price: float
    currency: str
    effective_from: datetime
    effective_to: datetime | None
    status: str  # draft | published | retired


@dataclass(frozen=True)
class EstimateRecord:
    estimate_id: str
    message_id: str
    provider_id: str
    country_code: str
    carrier: str
    estimated_cost: float
    currency: str
    rate_id: str
    rate_version: int
    as_of: datetime
    created_at: datetime


@dataclass(frozen=True)
class ActualCostRecord:
    actual_cost_id: str
    message_id: str
    provider_id: str
    actual_cost: float
    currency: str
    callback_state: str
    recorded_at: datetime
    provider_event_id: str
    idempotency_key: str
    idempotent_replay: bool


_ESTIMATES: dict[str, EstimateRecord] = {}
_ACTUAL_COSTS: dict[str, ActualCostRecord] = {}


def _rate_from_row(
    row: sa.Row,
    *,
    provider_code: str,
    carrier_code: str,
) -> Rate:
    return Rate(
        rate_id=f"rate_{row.id}",
        rate_version=1,
        provider_id=provider_code,
        country_code=row.country_code,
        carrier=carrier_code,
        unit_price=float(row.price_per_sms),
        currency=row.currency,
        effective_from=row.effective_from,
        effective_to=row.effective_to,
        status="published",
    )


async def find_rate(
    provider_id: str, country_code: str, carrier: str, as_of: datetime
) -> Rate | None:
    """Find published rate matching provider/country/carrier active at as_of."""
    as_of_utc = as_utc(as_of)
    async with get_session() as session:
        provider_pk = await fetch_provider_id_by_code(session, provider_id)
        if provider_pk is None:
            return None
        carrier_pk = await fetch_carrier_id_by_code(session, carrier)
        if carrier_pk is None:
            return None
        row = await fetch_active_rate(
            session,
            provider_pk,
            country_code,
            carrier_pk,
            as_of_utc,
        )
    if row is None:
        return None
    return _rate_from_row(row, provider_code=provider_id, carrier_code=carrier)


async def estimate_cost(
    message_id: str,
    provider_id: str,
    country_code: str,
    carrier: str,
    as_of: datetime,
) -> EstimateRecord | None:
    """Look up rate and create estimate. Returns None if no matching rate."""
    rate = await find_rate(provider_id, country_code, carrier, as_of)
    if rate is None:
        return None
    created = datetime.now(timezone.utc)
    estimate_id = str(uuid.uuid4())
    record = EstimateRecord(
        estimate_id=estimate_id,
        message_id=message_id,
        provider_id=provider_id,
        country_code=country_code,
        carrier=carrier,
        estimated_cost=rate.unit_price,
        currency=rate.currency,
        rate_id=rate.rate_id,
        rate_version=rate.rate_version,
        as_of=as_utc(as_of),
        created_at=created,
    )
    _ESTIMATES[estimate_id] = record
    return record


async def estimate_cost_batch(
    provider_ids: list[str],
    country_code: str,
    carrier: str,
    as_of: datetime,
) -> list[EstimateRecord]:
    """Estimate cost for each provider_id. Skip providers with no matching rate."""
    out: list[EstimateRecord] = []
    for pid in provider_ids:
        rec = await estimate_cost("", pid, country_code, carrier, as_of)
        if rec is not None:
            out.append(rec)
    return out


def record_actual_cost(
    message_id: str,
    provider_id: str,
    provider_event_id: str,
    idempotency_key: str,
    actual_cost: float,
    currency: str,
    callback_state: str,
    recorded_at: datetime,
) -> ActualCostRecord:
    """Record actual cost idempotently. Return existing record if idempotency_key already exists."""
    if idempotency_key in _ACTUAL_COSTS:
        existing = _ACTUAL_COSTS[idempotency_key]
        return ActualCostRecord(
            actual_cost_id=existing.actual_cost_id,
            message_id=existing.message_id,
            provider_id=existing.provider_id,
            actual_cost=existing.actual_cost,
            currency=existing.currency,
            callback_state=existing.callback_state,
            recorded_at=existing.recorded_at,
            provider_event_id=existing.provider_event_id,
            idempotency_key=existing.idempotency_key,
            idempotent_replay=True,
        )
    aid = str(uuid.uuid4())
    rec = ActualCostRecord(
        actual_cost_id=aid,
        message_id=message_id,
        provider_id=provider_id,
        actual_cost=actual_cost,
        currency=currency,
        callback_state=callback_state,
        recorded_at=as_utc(recorded_at),
        provider_event_id=provider_event_id,
        idempotency_key=idempotency_key,
        idempotent_replay=False,
    )
    _ACTUAL_COSTS[idempotency_key] = rec
    return rec


def clear() -> None:
    """Reset in-memory estimate and actual-cost stores (for tests)."""
    _ESTIMATES.clear()
    _ACTUAL_COSTS.clear()
