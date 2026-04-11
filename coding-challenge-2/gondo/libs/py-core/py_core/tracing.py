"""W3C trace context propagation and gRPC metadata helpers."""

from __future__ import annotations

import os
import uuid
from contextvars import ContextVar
from dataclasses import dataclass
from typing import Final

TRACEPARENT_HEADER: Final[str] = "traceparent"
REQUEST_ID_HEADER: Final[str] = "x-request-id"

_TRACEPARENT_VERSION: Final[str] = "00"
_TRACEPARENT_FLAGS_SAMPLED: Final[str] = "01"


@dataclass(frozen=True, slots=True)
class TraceContext:
    trace_id: str
    span_id: str
    parent_span_id: str | None = None


_current_trace: ContextVar[TraceContext | None] = ContextVar("_current_trace", default=None)


def get_current_trace_context() -> TraceContext | None:
    """Return the trace context for the current task, or None if unset."""
    return _current_trace.get()


def set_current_trace_context(ctx: TraceContext | None) -> None:
    """Bind *ctx* as the current trace context for this task (or clear with None)."""
    _current_trace.set(ctx)


def _is_fixed_hex(s: str, expected_len: int) -> bool:
    if len(s) != expected_len:
        return False
    try:
        int(s, 16)
    except ValueError:
        return False
    return True


def new_trace_context(*, parent: TraceContext | None = None) -> TraceContext:
    """Create a new root trace or child span context.

    Root: generates random trace_id + span_id.
    Child: inherits parent's trace_id, generates new span_id, sets parent_span_id.
    """
    if parent is None:
        trace_id = uuid.uuid4().hex
        span_id = os.urandom(8).hex()
        return TraceContext(trace_id=trace_id, span_id=span_id, parent_span_id=None)
    span_id = os.urandom(8).hex()
    return TraceContext(
        trace_id=parent.trace_id,
        span_id=span_id,
        parent_span_id=parent.span_id,
    )


def parse_traceparent(value: str) -> TraceContext | None:
    """Parse W3C traceparent header (version-traceid-parentid-flags).

    Returns None if invalid format.
    Example: '00-4bf92f3577b6a27610be46d4cac10152-00f067aa0ba902b7-01'
    """
    parts = value.split("-")
    if len(parts) != 4:
        return None
    version, trace_id_raw, span_id_raw, flags_raw = parts
    if version != _TRACEPARENT_VERSION:
        return None
    if not _is_fixed_hex(trace_id_raw, 32):
        return None
    if not _is_fixed_hex(span_id_raw, 16):
        return None
    if not _is_fixed_hex(flags_raw, 2):
        return None
    trace_id = trace_id_raw.lower()
    span_id = span_id_raw.lower()
    return TraceContext(trace_id=trace_id, span_id=span_id, parent_span_id=None)


def format_traceparent(ctx: TraceContext) -> str:
    """Format TraceContext as W3C traceparent string."""
    return f"{_TRACEPARENT_VERSION}-{ctx.trace_id}-{ctx.span_id}-{_TRACEPARENT_FLAGS_SAMPLED}"


def metadata_to_trace_context(md: tuple[tuple[str, str], ...]) -> TraceContext | None:
    """Extract trace context from gRPC invocation metadata (server side).

    Looks for 'traceparent' key in metadata tuples.
    """
    for key, val in md:
        if key == TRACEPARENT_HEADER:
            return parse_traceparent(val)
    return None


def trace_context_to_metadata(ctx: TraceContext) -> tuple[tuple[str, str], ...]:
    """Produce gRPC client metadata tuples from trace context."""
    return ((TRACEPARENT_HEADER, format_traceparent(ctx)),)
