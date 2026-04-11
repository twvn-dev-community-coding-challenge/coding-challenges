"""Tests for the shared FastAPI app factory."""

from __future__ import annotations

from fastapi import APIRouter
from fastapi.testclient import TestClient

from py_core.app import create_app


def test_create_app_returns_fastapi_with_title() -> None:
    app = create_app(title="Test Service", description="desc")
    assert app.title == "Test Service"


def test_create_app_includes_health_endpoint() -> None:
    app = create_app(title="Svc", description="d")
    client = TestClient(app)
    resp = client.get("/health")
    assert resp.status_code == 200
    assert resp.json() == {"status": "ok"}


def test_create_app_registers_custom_routers() -> None:
    router = APIRouter()

    @router.get("/custom")
    def _custom() -> dict[str, str]:
        return {"hello": "world"}

    app = create_app(title="Svc", description="d", routers=[router])
    client = TestClient(app)
    resp = client.get("/custom")
    assert resp.status_code == 200
    assert resp.json() == {"hello": "world"}


def test_create_app_passes_lifespan() -> None:
    started = []

    from contextlib import asynccontextmanager

    @asynccontextmanager
    async def _lifespan(_app: object):
        started.append(True)
        yield

    app = create_app(title="Svc", description="d", lifespan=_lifespan)
    with TestClient(app):
        assert started == [True]
