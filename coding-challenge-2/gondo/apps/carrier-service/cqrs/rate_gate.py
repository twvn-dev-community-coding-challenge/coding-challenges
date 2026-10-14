"""Limit concurrent outbound calls to provider/carrier HTTP endpoints (skeleton)."""

from __future__ import annotations

import asyncio
import os

_max = max(1, int(os.environ.get("CARRIER_MAX_CONCURRENT_SENDS", "4")))
_outbound_sem = asyncio.Semaphore(_max)


async def acquire_send_slot() -> None:
    await _outbound_sem.acquire()


def release_send_slot() -> None:
    _outbound_sem.release()
