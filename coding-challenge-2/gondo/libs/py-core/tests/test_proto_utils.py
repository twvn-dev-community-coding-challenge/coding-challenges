"""Tests for proto datetime utilities."""

from __future__ import annotations

from datetime import datetime, timezone

from py_core.proto_utils import as_utc, current_timestamp, datetime_to_timestamp


def test_as_utc_naive_gets_utc() -> None:
    naive = datetime(2026, 1, 15, 12, 0, 0)
    result = as_utc(naive)
    assert result.tzinfo == timezone.utc
    assert result.hour == 12


def test_as_utc_aware_converts() -> None:
    from datetime import timedelta

    tz_plus7 = timezone(timedelta(hours=7))
    aware = datetime(2026, 1, 15, 19, 0, 0, tzinfo=tz_plus7)
    result = as_utc(aware)
    assert result.tzinfo == timezone.utc
    assert result.hour == 12


def test_datetime_to_timestamp_roundtrips() -> None:
    dt = datetime(2026, 4, 11, 10, 30, 0, tzinfo=timezone.utc)
    ts = datetime_to_timestamp(dt)
    roundtripped = ts.ToDatetime()
    assert roundtripped.replace(tzinfo=timezone.utc) == dt


def test_current_timestamp_is_recent() -> None:
    ts = current_timestamp()
    dt = ts.ToDatetime().replace(tzinfo=timezone.utc)
    now = datetime.now(timezone.utc)
    assert abs((now - dt).total_seconds()) < 5
