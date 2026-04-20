"""In-memory SMS pipeline timeline per notification (runtime aggregation for UI / debugging)."""

from __future__ import annotations

import threading
from collections import defaultdict, deque
from datetime import datetime, timezone
from typing import Any

_MAX_EVENTS_PER_NOTIFICATION = 500

_lock = threading.Lock()
_buffers: dict[str, deque[dict[str, Any]]] = defaultdict(
    lambda: deque(maxlen=_MAX_EVENTS_PER_NOTIFICATION),
)


def clear() -> None:
    """Clear all runtime pipeline buffers (tests / store reset)."""
    with _lock:
        _buffers.clear()


def pipe(notification_id: str, phase: str, **detail: object) -> None:
    """Record one pipeline step; omits keys whose value is ``None``."""
    payload = {k: v for k, v in detail.items() if v is not None}
    append_pipeline_event(notification_id, phase, payload)


def append_pipeline_event(
    notification_id: str,
    phase: str,
    detail: dict[str, Any] | None = None,
) -> None:
    """Record one pipeline step for *notification_id* (thread-safe)."""
    if not notification_id.strip():
        return
    row: dict[str, Any] = {
        "timestamp": datetime.now(timezone.utc).isoformat(),
        "phase": phase,
        "detail": dict(detail) if detail else {},
    }
    with _lock:
        _buffers[notification_id].append(row)


def list_pipeline_events(notification_id: str) -> list[dict[str, Any]]:
    """Return ordered timeline for UI (newest-last)."""
    with _lock:
        q = _buffers.get(notification_id)
        if not q:
            return []
        return list(q)
