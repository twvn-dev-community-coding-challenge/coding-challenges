"""In-memory provider registry and routing rules for provider-service."""

from __future__ import annotations

import hashlib
from dataclasses import dataclass
from datetime import datetime, timezone
from typing import Final

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


_EFFECTIVE_START: Final[datetime] = datetime(2026, 1, 1, 0, 0, 0, tzinfo=timezone.utc)

_PROVIDERS_BY_ID: Final[dict[str, Provider]] = {
    "prv_01": Provider(
        provider_id="prv_01",
        provider_code="TWILIO",
        display_name="Twilio",
        status="active",
    ),
    "prv_02": Provider(
        provider_id="prv_02",
        provider_code="VONAGE",
        display_name="Vonage",
        status="active",
    ),
    "prv_03": Provider(
        provider_id="prv_03",
        provider_code="SINCH",
        display_name="Sinch",
        status="active",
    ),
}

_ROUTING_RULES: Final[tuple[RoutingRule, ...]] = (
    RoutingRule(
        routing_rule_id="rr_01",
        routing_rule_version=1,
        country_code="VN",
        carrier="VIETTEL",
        provider_id="prv_02",
        precedence=100,
        effective_from=_EFFECTIVE_START,
        effective_to=None,
        status="published",
    ),
    RoutingRule(
        routing_rule_id="rr_02",
        routing_rule_version=1,
        country_code="VN",
        carrier="VIETTEL",
        provider_id="prv_01",
        precedence=90,
        effective_from=_EFFECTIVE_START,
        effective_to=None,
        status="published",
    ),
    RoutingRule(
        routing_rule_id="rr_03",
        routing_rule_version=1,
        country_code="VN",
        carrier="MOBIFONE",
        provider_id="prv_01",
        precedence=100,
        effective_from=_EFFECTIVE_START,
        effective_to=None,
        status="published",
    ),
    RoutingRule(
        routing_rule_id="rr_04",
        routing_rule_version=1,
        country_code="US",
        carrier="T-MOBILE",
        provider_id="prv_01",
        precedence=100,
        effective_from=_EFFECTIVE_START,
        effective_to=None,
        status="published",
    ),
    RoutingRule(
        routing_rule_id="rr_05",
        routing_rule_version=1,
        country_code="US",
        carrier="T-MOBILE",
        provider_id="prv_03",
        precedence=80,
        effective_from=_EFFECTIVE_START,
        effective_to=None,
        status="published",
    ),
    RoutingRule(
        routing_rule_id="rr_06",
        routing_rule_version=1,
        country_code="US",
        carrier="AT&T",
        provider_id="prv_02",
        precedence=100,
        effective_from=_EFFECTIVE_START,
        effective_to=None,
        status="published",
    ),
)


def _normalize(s: str) -> str:
    return s.strip()


def _is_active_at(rule: RoutingRule, as_of: datetime) -> bool:
    if rule.status != "published":
        return False
    if as_of < rule.effective_from:
        return False
    if rule.effective_to is not None and as_of >= rule.effective_to:
        return False
    return True


def get_provider(provider_id: str) -> Provider | None:
    """Look up provider by ID."""
    return _PROVIDERS_BY_ID.get(provider_id)


def resolve_routing(country_code: str, carrier: str, as_of: datetime) -> list[RoutingCandidate]:
    """Filter published rules matching (country_code, carrier) active at as_of."""
    cc = _normalize(country_code).upper()
    cr = _normalize(carrier)

    matched: list[tuple[RoutingRule, Provider]] = []
    for rule in _ROUTING_RULES:
        if _normalize(rule.country_code).upper() != cc:
            continue
        if _normalize(rule.carrier) != cr:
            continue
        if not _is_active_at(rule, as_of):
            continue
        provider = _PROVIDERS_BY_ID.get(rule.provider_id)
        if provider is None:
            continue
        matched.append((rule, provider))

    matched.sort(key=lambda t: t[0].precedence, reverse=True)

    resolved: list[RoutingCandidate] = []
    for rule, provider in matched:
        resolved.append(
            RoutingCandidate(
                provider_id=provider.provider_id,
                provider_code=provider.provider_code,
                routing_rule_id=rule.routing_rule_id,
                routing_rule_version=rule.routing_rule_version,
                effective_from=rule.effective_from,
                effective_to=rule.effective_to,
                resolved_at=as_of,
                precedence=rule.precedence,
            )
        )
    return resolved


def _healthy_candidates(candidates: list[RoutingCandidate]) -> list[RoutingCandidate]:
    out: list[RoutingCandidate] = []
    for c in candidates:
        p = _PROVIDERS_BY_ID.get(c.provider_id)
        if p is not None and p.status == "active":
            out.append(c)
    return out


def _normalize_policy(policy: str) -> str:
    p = policy.strip().lower()
    if p == "":
        return "highest_precedence"
    return p


def _round_robin_pick(candidates: list[RoutingCandidate], message_id: str) -> RoutingCandidate:
    n = len(candidates)
    if n == 1:
        return candidates[0]
    key = message_id if message_id else "|".join(c.provider_id for c in candidates)
    digest = hashlib.sha256(key.encode("utf-8")).hexdigest()
    idx = int(digest[:16], 16) % n
    return candidates[idx]


def select_provider(
    country_code: str,
    carrier: str,
    as_of: datetime,
    policy: str,
    *,
    message_id: str = "",
) -> SelectionResult | None:
    """Resolve candidates, filter healthy (status=active), select by policy."""
    candidates = resolve_routing(country_code, carrier, as_of)
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
