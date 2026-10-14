"""Tests for W3C trace context and gRPC metadata helpers."""

from __future__ import annotations

import re

from py_core.tracing import (
    TRACEPARENT_HEADER,
    TraceContext,
    format_traceparent,
    get_current_trace_context,
    metadata_to_trace_context,
    new_trace_context,
    parse_traceparent,
    set_current_trace_context,
    trace_context_to_metadata,
)

_HEX32 = re.compile(r"^[0-9a-f]{32}$")
_HEX16 = re.compile(r"^[0-9a-f]{16}$")

_VALID_TRACEPARENT = "00-4bf92f3577b6a27610be46d4cac10152-00f067aa0ba902b7-01"


def test_new_trace_context_root() -> None:
    ctx = new_trace_context()
    assert _HEX32.match(ctx.trace_id)
    assert _HEX16.match(ctx.span_id)
    assert ctx.parent_span_id is None


def test_new_trace_context_child() -> None:
    root = new_trace_context()
    child = new_trace_context(parent=root)
    assert child.trace_id == root.trace_id
    assert child.span_id != root.span_id
    assert _HEX16.match(child.span_id)
    assert child.parent_span_id == root.span_id


def test_parse_traceparent_valid() -> None:
    parsed = parse_traceparent(_VALID_TRACEPARENT)
    assert parsed is not None
    assert parsed.trace_id == "4bf92f3577b6a27610be46d4cac10152"
    assert parsed.span_id == "00f067aa0ba902b7"
    assert parsed.parent_span_id is None


def test_parse_traceparent_invalid() -> None:
    assert parse_traceparent("") is None
    assert parse_traceparent("not-a-traceparent") is None
    assert parse_traceparent("01-4bf92f3577b6a27610be46d4cac10152-00f067aa0ba902b7-01") is None


def test_format_traceparent_roundtrip() -> None:
    ctx = TraceContext(
        trace_id="4bf92f3577b6a27610be46d4cac10152",
        span_id="00f067aa0ba902b7",
        parent_span_id="aaaaaaaaaaaaaaaa",
    )
    formatted = format_traceparent(ctx)
    parsed = parse_traceparent(formatted)
    assert parsed is not None
    assert parsed.trace_id == ctx.trace_id
    assert parsed.span_id == ctx.span_id


def test_get_set_current_trace_context() -> None:
    ctx = new_trace_context()
    try:
        set_current_trace_context(ctx)
        assert get_current_trace_context() == ctx
    finally:
        set_current_trace_context(None)


def test_get_current_trace_context_default_none() -> None:
    assert get_current_trace_context() is None


def test_metadata_to_trace_context() -> None:
    md = (
        ("traceparent", _VALID_TRACEPARENT),
        ("other", "value"),
    )
    ctx = metadata_to_trace_context(md)
    assert ctx is not None
    assert ctx.trace_id == "4bf92f3577b6a27610be46d4cac10152"
    assert ctx.span_id == "00f067aa0ba902b7"


def test_metadata_to_trace_context_missing() -> None:
    md = (("other", "value"),)
    assert metadata_to_trace_context(md) is None


def test_trace_context_to_metadata() -> None:
    ctx = TraceContext(
        trace_id="4bf92f3577b6a27610be46d4cac10152",
        span_id="00f067aa0ba902b7",
    )
    md = trace_context_to_metadata(ctx)
    assert md == (("traceparent", format_traceparent(ctx)),)
    assert md[0][0] == TRACEPARENT_HEADER
