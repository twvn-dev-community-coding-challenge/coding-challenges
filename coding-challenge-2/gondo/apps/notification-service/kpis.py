"""SMS KPI read model — aggregates **in-memory** notifications (User Story 5).

Costs and states are summed from ``store.list_notifications()``; resets when the process
or tests call ``store.clear()`` — same semantics as the rest of notification-service.
"""

from __future__ import annotations

from collections import defaultdict
from typing import Any

from models import Notification
from store import list_notifications


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


def build_sms_kpis() -> dict[str, Any]:
    """Return rolled-up KPIs for observability (challenge US5)."""
    items = list_notifications()

    by_provider: dict[str, dict[str, Any]] = defaultdict(_empty_bucket)
    by_country: dict[str, dict[str, Any]] = defaultdict(_empty_bucket)
    overall = _empty_bucket()

    for n in items:
        pid = n.selected_provider_id or "unassigned"
        cc = (n.channel_payload.get("country_code") or "").strip().upper() or "unknown"

        _accumulate(overall, n)
        _accumulate(by_provider[pid], n)
        _accumulate(by_country[cc], n)

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

    return {
        "source": "in_memory_notifications",
        "currency_note": (
            "Costs are summed as floats; production would normalize currency per charge record."
        ),
        "overall": ov,
        "by_provider": providers_sorted,
        "by_country": countries_sorted,
    }
