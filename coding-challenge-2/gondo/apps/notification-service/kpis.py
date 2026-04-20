"""SMS KPI read model — aggregates **in-memory** notifications (User Story 5).

Costs and states are summed from ``store.list_notifications()``; resets when the process
or tests call ``store.clear()`` — same semantics as the rest of notification-service.
"""

from __future__ import annotations

from collections import defaultdict
from datetime import datetime, timezone
from typing import Any

from models import Notification
from store import list_notifications


def parse_iso_datetime(value: str) -> datetime:
    """Parse ISO 8601 into **UTC** ``datetime`` (naive input treated as UTC)."""
    s = value.strip()
    if s.endswith("Z"):
        s = s[:-1] + "+00:00"
    dt = datetime.fromisoformat(s)
    if dt.tzinfo is None:
        return dt.replace(tzinfo=timezone.utc)
    return dt.astimezone(timezone.utc)


def _utc(naive_or_aware: datetime) -> datetime:
    if naive_or_aware.tzinfo is None:
        return naive_or_aware.replace(tzinfo=timezone.utc)
    return naive_or_aware.astimezone(timezone.utc)


def _created_at_in_window(
    n: Notification,
    window_from: datetime | None,
    window_to: datetime | None,
) -> bool:
    if window_from is None and window_to is None:
        return True
    ts = _utc(n.created_at)
    if window_from is not None and ts < window_from:
        return False
    if window_to is not None and ts > window_to:
        return False
    return True


def _empty_bucket() -> dict[str, Any]:
    return {
        "volume": 0,
        "total_estimated_cost": 0.0,
        "total_actual_cost": 0.0,
        "with_estimate_count": 0,
        "with_actual_count": 0,
        "send_success": 0,
        "send_failed": 0,
        "carrier_rejected": 0,
        "in_flight": 0,
    }


def _accumulate(bucket: dict[str, Any], n: Notification) -> None:
    bucket["volume"] += 1
    if n.estimated_cost is not None:
        bucket["total_estimated_cost"] += float(n.estimated_cost)
        bucket["with_estimate_count"] += 1
    if n.last_actual_cost is not None:
        bucket["total_actual_cost"] += float(n.last_actual_cost)
        bucket["with_actual_count"] += 1
    st = n.state
    if st == "Send-success":
        bucket["send_success"] += 1
    elif st == "Send-failed":
        bucket["send_failed"] += 1
    elif st == "Carrier-rejected":
        bucket["carrier_rejected"] += 1
    else:
        bucket["in_flight"] += 1


def _terminal_success_rate(success: int, failure: int) -> float | None:
    terminal = success + failure
    if terminal == 0:
        return None
    return round(success / terminal, 6)


def _terminal_failure_rate(success: int, failure: int) -> float | None:
    terminal = success + failure
    if terminal == 0:
        return None
    return round(failure / terminal, 6)


def build_sms_kpis(
    *,
    created_from: datetime | None = None,
    created_to: datetime | None = None,
) -> dict[str, Any]:
    """Return rolled-up KPIs for observability (challenge US5).

    Optional ``created_from`` / ``created_to`` bound ``Notification.created_at``
    (inclusive, compared in UTC).
    """
    window_from = _utc(created_from) if created_from is not None else None
    window_to = _utc(created_to) if created_to is not None else None

    items = [
        n
        for n in list_notifications()
        if _created_at_in_window(n, window_from, window_to)
    ]

    by_provider: dict[str, dict[str, Any]] = defaultdict(_empty_bucket)
    by_country: dict[str, dict[str, Any]] = defaultdict(_empty_bucket)
    by_calling_domain: dict[str, dict[str, Any]] = defaultdict(_empty_bucket)
    overall = _empty_bucket()

    for n in items:
        pid = n.selected_provider_id or "unassigned"
        cc = (n.channel_payload.get("country_code") or "").strip().upper() or "unknown"
        cd_raw = (n.channel_payload.get("calling_domain") or "").strip()
        cd = cd_raw if cd_raw else "unattributed"

        _accumulate(overall, n)
        _accumulate(by_provider[pid], n)
        _accumulate(by_country[cc], n)
        _accumulate(by_calling_domain[cd], n)

    def finalize_row(b: dict[str, Any]) -> dict[str, Any]:
        succ = int(b["send_success"])
        fail = int(b["send_failed"]) + int(b["carrier_rejected"])
        out = dict(b)
        out["terminal_failure"] = fail
        out["terminal_success_rate"] = _terminal_success_rate(succ, fail)
        out["terminal_failure_rate"] = _terminal_failure_rate(succ, fail)
        return out

    ov = finalize_row(dict(overall))
    ov["total_notifications"] = len(items)

    providers_sorted = sorted(
        (
            {"provider_id": k, **finalize_row(dict(v))}
            for k, v in by_provider.items()
        ),
        key=lambda x: (-x["volume"], x["provider_id"]),
    )
    countries_sorted = sorted(
        (
            {"country_code": k, **finalize_row(dict(v))}
            for k, v in by_country.items()
        ),
        key=lambda x: (-x["volume"], x["country_code"]),
    )
    calling_domains_sorted = sorted(
        (
            {"calling_domain": k, **finalize_row(dict(v))}
            for k, v in by_calling_domain.items()
        ),
        key=lambda x: (-x["volume"], x["calling_domain"]),
    )

    return {
        "source": "in_memory_notifications",
        "created_at_filter": {
            "from": window_from.isoformat() if window_from is not None else None,
            "to": window_to.isoformat() if window_to is not None else None,
        },
        "currency_note": (
            "Costs are summed as floats; production would normalize currency per charge record."
        ),
        "overall": ov,
        "by_provider": providers_sorted,
        "by_country": countries_sorted,
        "by_calling_domain": calling_domains_sorted,
    }
