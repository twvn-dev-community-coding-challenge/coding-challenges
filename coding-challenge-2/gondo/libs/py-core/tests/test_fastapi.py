"""Tests for the shared FastAPI app factory."""

from __future__ import annotations

import logging

import pytest
from fastapi import APIRouter
from fastapi.testclient import TestClient

from py_core.app import create_app
from py_core.redact import DEFAULT_POLICY, RedactionPolicy, RedactionRule
from py_core.tracing import get_current_trace_context


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


def test_create_app_tracing_middleware_creates_trace_context() -> None:
    app = create_app(title="Svc", description="d")
    client = TestClient(app)
    resp = client.get("/health")
    assert resp.status_code == 200


def test_create_app_tracing_middleware_parses_traceparent() -> None:
    captured: dict[str, str] = {}
    router = APIRouter()

    @router.get("/check-trace")
    def check_trace() -> dict[str, bool]:
        ctx = get_current_trace_context()
        if ctx:
            captured["trace_id"] = ctx.trace_id
            captured["span_id"] = ctx.span_id
        return {"ok": True}

    app = create_app(title="Svc", description="d", routers=[router])
    client = TestClient(app)
    resp = client.get(
        "/check-trace",
        headers={"traceparent": "00-4bf92f3577b6a27610be46d4cac10152-00f067aa0ba902b7-01"},
    )
    assert resp.status_code == 200
    assert captured["trace_id"] == "4bf92f3577b6a27610be46d4cac10152"


def test_create_app_tracing_disabled() -> None:
    captured: dict[str, str | None] = {}
    router = APIRouter()

    @router.get("/check-trace")
    def check_trace() -> dict[str, bool]:
        ctx = get_current_trace_context()
        captured["ctx"] = ctx.trace_id if ctx else None
        return {"ok": True}

    app = create_app(
        title="Svc",
        description="d",
        routers=[router],
        enable_request_tracing=False,
    )
    client = TestClient(app)
    resp = client.get(
        "/check-trace",
        headers={"traceparent": "00-4bf92f3577b6a27610be46d4cac10152-00f067aa0ba902b7-01"},
    )
    assert resp.status_code == 200
    assert captured["ctx"] is None


def test_create_app_service_name_parameter() -> None:
    app = create_app(
        title="Svc",
        description="d",
        service_name="my-service",
    )
    assert app.title == "Svc"


def test_payload_logging_logs_request_body(caplog: pytest.LogCaptureFixture) -> None:
    router = APIRouter()

    @router.post("/echo")
    def echo() -> dict[str, bool]:
        return {"ok": True}

    app = create_app(title="T", description="d", enable_payload_logging=True, routers=[router])
    client = TestClient(app)

    with caplog.at_level(logging.DEBUG, logger="py_core.payload.http"):
        client.post("/echo", json={"phone_number": "+84912345678", "name": "John"})

    request_logs = [
        r
        for r in caplog.records
        if r.name == "py_core.payload.http" and "http_request" in r.getMessage()
    ]
    assert len(request_logs) >= 1
    body = getattr(request_logs[0], "body", None)
    assert body is not None
    assert isinstance(body, dict)
    assert "+84912345678" not in str(body)
    assert body.get("name") == "John"

    response_logs = [
        r
        for r in caplog.records
        if r.name == "py_core.payload.http" and "http_response" in r.getMessage()
    ]
    assert len(response_logs) >= 1
    assert getattr(response_logs[0], "status_code", None) == 200


def test_payload_logging_disabled_by_default(
    caplog: pytest.LogCaptureFixture,
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    monkeypatch.delenv("LOG_PAYLOADS", raising=False)

    router = APIRouter()

    @router.post("/echo")
    def echo() -> dict[str, bool]:
        return {"ok": True}

    app = create_app(title="T", description="d", routers=[router])
    client = TestClient(app)

    with caplog.at_level(logging.DEBUG, logger="py_core.payload.http"):
        client.post("/echo", json={"phone_number": "+84912345678"})

    payload_records = [r for r in caplog.records if r.name == "py_core.payload.http"]
    assert payload_records == []


def test_payload_logging_env_var(
    caplog: pytest.LogCaptureFixture,
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    monkeypatch.setenv("LOG_PAYLOADS", "true")

    router = APIRouter()

    @router.post("/echo")
    def echo() -> dict[str, bool]:
        return {"ok": True}

    app = create_app(title="T", description="d", routers=[router])
    client = TestClient(app)

    with caplog.at_level(logging.DEBUG, logger="py_core.payload.http"):
        client.post("/echo", json={"x": 1})

    request_logs = [
        r
        for r in caplog.records
        if r.name == "py_core.payload.http" and "http_request" in r.getMessage()
    ]
    assert len(request_logs) >= 1


def test_payload_logging_custom_policy(caplog: pytest.LogCaptureFixture) -> None:
    custom = RedactionPolicy(
        rules=(
            *DEFAULT_POLICY.rules,
            RedactionRule(field_pattern="custom_secret", strategy="redact"),
        ),
    )

    router = APIRouter()

    @router.post("/echo")
    def echo() -> dict[str, bool]:
        return {"ok": True}

    app = create_app(
        title="T",
        description="d",
        enable_payload_logging=True,
        payload_redaction_policy=custom,
        routers=[router],
    )
    client = TestClient(app)

    with caplog.at_level(logging.DEBUG, logger="py_core.payload.http"):
        client.post("/echo", json={"custom_secret": "hide-me", "visible": "ok"})

    request_logs = [
        r
        for r in caplog.records
        if r.name == "py_core.payload.http" and "http_request" in r.getMessage()
    ]
    assert len(request_logs) >= 1
    body = getattr(request_logs[0], "body", None)
    assert body is not None
    assert body.get("custom_secret") == "[REDACTED]"
    assert body.get("visible") == "ok"
