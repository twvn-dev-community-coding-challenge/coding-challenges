"""HTTP proxy to NATS: publish logical topics without embedding the client SDK in tools."""

from __future__ import annotations

import json
import os
from typing import Any

import nats
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field

from py_core.bus.topics import topic_to_subject

_nc: object | None = None

app = FastAPI(
    title="Message Bus Gateway",
    description="Sidecar-style HTTP → NATS publish (infra; broker details stay in ops).",
)


class PublishBody(BaseModel):
    topic: str = Field(..., description="Logical topic, e.g. sms.dispatch.requested")
    payload: dict[str, Any] = Field(default_factory=dict)


@app.on_event("startup")
async def startup() -> None:
    global _nc
    raw = os.environ.get("NATS_URL", "nats://127.0.0.1:4222")
    servers = [s.strip() for s in raw.split(",") if s.strip()]
    _nc = await nats.connect(servers=servers)


@app.on_event("shutdown")
async def shutdown() -> None:
    global _nc
    if _nc is not None:
        await _nc.drain()
        await _nc.close()
        _nc = None


@app.post("/v1/publish")
async def publish_message(body: PublishBody) -> dict[str, str]:
    if _nc is None:
        raise HTTPException(status_code=503, detail="nats_not_connected")
    subject = topic_to_subject(body.topic.strip())
    data = json.dumps(body.payload).encode("utf-8")
    await _nc.publish(subject, data)
    return {"status": "published", "subject": subject}


@app.get("/health")
def health() -> dict[str, str]:
    return {"status": "ok"}
