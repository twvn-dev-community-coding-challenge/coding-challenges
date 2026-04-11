"""In-memory rate table, estimates, and actual cost recording."""

from __future__ import annotations

import uuid
from dataclasses import dataclass
from datetime import datetime, timezone
from typing import Final


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


_EFFECTIVE_FROM: Final[datetime] = datetime(2026, 1, 1, tzinfo=timezone.utc)

_RATES: tuple[Rate, ...] = (
    Rate(
        rate_id="rate_101",
        rate_version=1,
        provider_id="prv_01",
        country_code="VN",
        carrier="VIETTEL",
        unit_price=0.015,
        currency="USD",
        effective_from=_EFFECTIVE_FROM,
        effective_to=None,
        status="published",
    ),
    Rate(
        rate_id="rate_102",
        rate_version=1,
        provider_id="prv_02",
        country_code="VN",
        carrier="VIETTEL",
        unit_price=0.018,
        currency="USD",
        effective_from=_EFFECTIVE_FROM,
        effective_to=None,
        status="published",
    ),
    Rate(
        rate_id="rate_103",
        rate_version=1,
        provider_id="prv_01",
        country_code="VN",
        carrier="MOBIFONE",
        unit_price=0.016,
        currency="USD",
        effective_from=_EFFECTIVE_FROM,
        effective_to=None,
        status="published",
    ),
    Rate(
        rate_id="rate_104",
        rate_version=1,
        provider_id="prv_01",
        country_code="US",
        carrier="T-MOBILE",
        unit_price=0.008,
        currency="USD",
        effective_from=_EFFECTIVE_FROM,
        effective_to=None,
        status="published",
    ),
    Rate(
        rate_id="rate_105",
        rate_version=1,
        provider_id="prv_03",
        country_code="US",
        carrier="T-MOBILE",
        unit_price=0.009,
        currency="USD",
        effective_from=_EFFECTIVE_FROM,
        effective_to=None,
        status="published",
    ),
    Rate(
        rate_id="rate_106",
        rate_version=1,
        provider_id="prv_02",
        country_code="US",
        carrier="AT&T",
        unit_price=0.010,
        currency="USD",
        effective_from=_EFFECTIVE_FROM,
        effective_to=None,
        status="published",
    ),
)

_ESTIMATES: dict[str, EstimateRecord] = {}
_ACTUAL_COSTS: dict[str, ActualCostRecord] = {}


def _as_utc(dt: datetime) -> datetime:
    if dt.tzinfo is None:
        return dt.replace(tzinfo=timezone.utc)
    return dt.astimezone(timezone.utc)


def _rate_active_at(rate: Rate, as_of: datetime) -> bool:
    t = _as_utc(as_of)
    if t < _as_utc(rate.effective_from):
        return False
    if rate.effective_to is not None and t > _as_utc(rate.effective_to):
        return False
    return rate.status == "published"


def find_rate(
    provider_id: str, country_code: str, carrier: str, as_of: datetime
) -> Rate | None:
    """Find published rate matching provider/country/carrier active at as_of."""
    for rate in _RATES:
        if (
            rate.provider_id == provider_id
            and rate.country_code == country_code
            and rate.carrier == carrier
            and _rate_active_at(rate, as_of)
        ):
            return rate
    return None


def estimate_cost(
    message_id: str,
    provider_id: str,
    country_code: str,
    carrier: str,
    as_of: datetime,
) -> EstimateRecord | None:
    """Look up rate and create estimate. Returns None if no matching rate."""
    rate = find_rate(provider_id, country_code, carrier, as_of)
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
        as_of=_as_utc(as_of),
        created_at=created,
    )
    _ESTIMATES[estimate_id] = record
    return record


def estimate_cost_batch(
    provider_ids: list[str],
    country_code: str,
    carrier: str,
    as_of: datetime,
) -> list[EstimateRecord]:
    """Estimate cost for each provider_id. Skip providers with no matching rate."""
    out: list[EstimateRecord] = []
    for pid in provider_ids:
        rec = estimate_cost("", pid, country_code, carrier, as_of)
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
        recorded_at=_as_utc(recorded_at),
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
