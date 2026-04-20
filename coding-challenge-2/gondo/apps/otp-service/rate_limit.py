"""Fixed-window rate limits per client IP for issue / verify endpoints."""

from __future__ import annotations

import time
from collections import defaultdict, deque
from typing import Final

_WINDOW_SECONDS: Final[int] = 60

_issue_events: dict[str, deque[float]] = defaultdict(deque)
_verify_events: dict[str, deque[float]] = defaultdict(deque)


def _prune_old(events: deque[float], now: float) -> None:
    cutoff = now - _WINDOW_SECONDS
    while events and events[0] < cutoff:
        events.popleft()


def allow_issue(client_ip: str, limit_per_window: int) -> bool:
    """Return False when *client_ip* has met or exceeded *limit_per_window* in the last minute."""
    if limit_per_window <= 0:
        return True
    now = time.monotonic()
    q = _issue_events[client_ip]
    _prune_old(q, now)
    if len(q) >= limit_per_window:
        return False
    q.append(now)
    return True


def allow_verify(client_ip: str, limit_per_window: int) -> bool:
    if limit_per_window <= 0:
        return True
    now = time.monotonic()
    q = _verify_events[client_ip]
    _prune_old(q, now)
    if len(q) >= limit_per_window:
        return False
    q.append(now)
    return True


def reset_for_testing() -> None:
    """Clear counters (pytest only)."""
    _issue_events.clear()
    _verify_events.clear()
