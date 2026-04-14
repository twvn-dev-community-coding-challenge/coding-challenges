"""Provider registry and routing rules backed by PostgreSQL (provider_service schema)."""

from __future__ import annotations

import hashlib
from collections.abc import Mapping
from dataclasses import dataclass
from datetime import datetime
from uuid import UUID

from py_core.db import get_session

from repository import (
    fetch_active_routing_rules,
    fetch_all_providers,
    fetch_provider_by_code,
)

# ---------------------------------------------------------------------------
# Data model (Provider TD Section 5)
# ---------------------------------------------------------------------------


@dataclass(frozen=True)
class Provider:
    provider_id: str
    provider_code: str
    display_name: str
    status: str  # active | deprecated | disabled


@dataclass(frozen=True)
class RoutingRule:
    routing_rule_id: str
    routing_rule_version: int
    country_code: str
    carrier: str
    provider_id: str
    precedence: int
    effective_from: datetime
    effective_to: datetime | None
    status: str  # draft | published | retired


@dataclass(frozen=True)
class RoutingCandidate:
    provider_id: str
    provider_code: str
    routing_rule_id: str
    routing_rule_version: int
    effective_from: datetime
    effective_to: datetime | None
    resolved_at: datetime
    precedence: int


@dataclass(frozen=True)
class SelectionResult:
    selected_provider_id: str
    selected_provider_code: str
    selection_policy: str
    selection_reason: str
    routing_rule_id: str
    routing_rule_version: int


def _provider_from_mapping(m: Mapping[str, object]) -> Provider:
    code = m.get("code")
    name = m.get("name")
    if not isinstance(code, str) or not isinstance(name, str):
        msg = "provider row must include string code and name"
        raise TypeError(msg)
    pcode = name.upper().replace(" ", "_")
    return Provider(
        provider_id=code,
        provider_code=pcode,
        display_name=name,
        status="active",
    )


def _routing_rule_id_from_row(rid: object) -> str:
    if isinstance(rid, UUID):
        return str(rid)
    if isinstance(rid, str):
        return rid
    msg = "routing rule row must include UUID or string id"
    raise TypeError(msg)


def _routing_candidate_from_rule_row(
    m: Mapping[str, object],
    provider: Provider,
    as_of: datetime,
) -> RoutingCandidate:
    rid = m.get("id")
    rrv = m.get("routing_rule_version")
    priority = m.get("priority")
    eff_from = m.get("effective_from")
    eff_to = m.get("effective_to")
    routing_rule_id = _routing_rule_id_from_row(rid)
    if not isinstance(rrv, int):
        msg = "routing rule row must include integer routing_rule_version"
        raise TypeError(msg)
    if not isinstance(priority, int):
        msg = "routing rule row must include integer priority"
        raise TypeError(msg)
    if not isinstance(eff_from, datetime):
        msg = "routing rule row must include datetime effective_from"
        raise TypeError(msg)
    if eff_to is not None and not isinstance(eff_to, datetime):
        msg = "routing rule effective_to must be datetime or None"
        raise TypeError(msg)

    return RoutingCandidate(
        provider_id=provider.provider_id,
        provider_code=provider.provider_code,
        routing_rule_id=routing_rule_id,
        routing_rule_version=rrv,
        effective_from=eff_from,
        effective_to=eff_to,
        resolved_at=as_of,
        precedence=priority,
    )


def _providers_by_int_id_from_rows(rows: list[object]) -> dict[int, Provider]:
    out: dict[int, Provider] = {}
    for row in rows:
        m = getattr(row, "_mapping", None)
        if not isinstance(m, Mapping):
            msg = "expected SQLAlchemy Row with _mapping"
            raise TypeError(msg)
        pid = m.get("id")
        if not isinstance(pid, int):
            msg = "provider row must include integer id"
            raise TypeError(msg)
        p = _provider_from_mapping(m)
        out[pid] = p
    return out


async def get_provider(provider_id: str) -> Provider | None:
    """Look up provider by business code (e.g. prv_01)."""
    async with get_session() as session:
        row = await fetch_provider_by_code(session, provider_id)
    if row is None:
        return None
    raw = getattr(row, "_mapping", None)
    if not isinstance(raw, Mapping):
        msg = "expected SQLAlchemy Row with _mapping"
        raise TypeError(msg)
    return _provider_from_mapping(raw)


async def resolve_routing(
    country_code: str, carrier: str, as_of: datetime
) -> list[RoutingCandidate]:
    """Filter rules matching (country_code, carrier) active at as_of; order by version DESC, priority DESC."""
    async with get_session() as session:
        rule_rows = await fetch_active_routing_rules(
            session, country_code, carrier, as_of
        )
        if not rule_rows:
            return []
        provider_rows = await fetch_all_providers(session)

    by_int_id = _providers_by_int_id_from_rows(provider_rows)
    resolved: list[RoutingCandidate] = []
    for rrow in rule_rows:
        raw = getattr(rrow, "_mapping", None)
        if not isinstance(raw, Mapping):
            continue
        m: Mapping[str, object] = raw
        pid_obj = m.get("provider_id")
        if not isinstance(pid_obj, int):
            continue
        provider = by_int_id.get(pid_obj)
        if provider is None:
            continue
        resolved.append(_routing_candidate_from_rule_row(m, provider, as_of))
    return resolved


def _healthy_candidates(candidates: list[RoutingCandidate]) -> list[RoutingCandidate]:
    """All providers present in DB are treated as healthy (no status column on providers)."""
    return list(candidates)


def _normalize_policy(policy: str) -> str:
    p = policy.strip().lower()
    if p == "":
        return "highest_precedence"
    return p


def _round_robin_pick(
    candidates: list[RoutingCandidate], message_id: str
) -> RoutingCandidate:
    n = len(candidates)
    if n == 1:
        return candidates[0]
    key = message_id if message_id else "|".join(c.provider_id for c in candidates)
    digest = hashlib.sha256(key.encode("utf-8")).hexdigest()
    idx = int(digest[:16], 16) % n
    return candidates[idx]


async def select_provider(
    country_code: str,
    carrier: str,
    as_of: datetime,
    policy: str,
    *,
    message_id: str = "",
) -> SelectionResult | None:
    """Resolve candidates, filter healthy, select by policy."""
    candidates = await resolve_routing(country_code, carrier, as_of)
    healthy = _healthy_candidates(candidates)
    if not healthy:
        return None

    pol = _normalize_policy(policy)
    if pol in ("highest_precedence", "lowest_cost_healthy"):
        chosen = healthy[0]
        reason = "highest_precedence_among_healthy"
        if pol == "lowest_cost_healthy":
            reason = "lowest_cost_delegated_to_precedence_until_charging"
    elif pol == "round_robin":
        chosen = _round_robin_pick(healthy, message_id)
        reason = "round_robin_by_message_id_hash"
    else:
        chosen = healthy[0]
        reason = f"unknown_policy_fallback_to_highest_precedence:{pol}"

    return SelectionResult(
        selected_provider_id=chosen.provider_id,
        selected_provider_code=chosen.provider_code,
        selection_policy=pol,
        selection_reason=reason,
        routing_rule_id=chosen.routing_rule_id,
        routing_rule_version=chosen.routing_rule_version,
    )
