"""Protobuf datetime conversion utilities shared across gRPC services."""

from __future__ import annotations

from datetime import datetime, timezone

from google.protobuf import timestamp_pb2


def as_utc(dt: datetime) -> datetime:
    """Ensure *dt* is UTC-aware; treat naive datetimes as UTC."""
    if dt.tzinfo is None:
        return dt.replace(tzinfo=timezone.utc)
    return dt.astimezone(timezone.utc)


def datetime_to_timestamp(dt: datetime) -> timestamp_pb2.Timestamp:
    """Convert a Python datetime to a Protobuf Timestamp."""
    ts = timestamp_pb2.Timestamp()
    ts.FromDatetime(as_utc(dt))
    return ts


def current_timestamp() -> timestamp_pb2.Timestamp:
    """Return a Protobuf Timestamp set to the current time."""
    ts = timestamp_pb2.Timestamp()
    ts.GetCurrentTime()
    return ts
