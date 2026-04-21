"""FastAPI app factory with shared /health endpoint."""

from __future__ import annotations

import hashlib
import json
import logging
import os
from collections.abc import AsyncGenerator, Callable
from typing import Any

from fastapi import APIRouter, FastAPI
from fastapi.middleware.cors import CORSMiddleware
from starlette.middleware.base import BaseHTTPMiddleware
from starlette.requests import Request

from py_core.logging import _should_log_payloads
from py_core.redact import DEFAULT_POLICY, RedactionPolicy
from py_core.tracing import (
    REQUEST_ID_HEADER,
    TRACEPARENT_HEADER,
    TraceContext,
    new_trace_context,
    parse_traceparent,
    set_current_trace_context,
)

health_router = APIRouter()


@health_router.get("/health")
def health() -> dict[str, str]:
    return {"status": "ok"}


def _trace_context_from_request_id(request_id: str) -> TraceContext:
    trace_id = hashlib.sha256(request_id.encode("utf-8")).hexdigest()[:32]
    span_id = os.urandom(8).hex()
    return TraceContext(trace_id=trace_id, span_id=span_id, parent_span_id=None)


def _resolve_trace_context_from_request(request: Request) -> TraceContext:
    traceparent_raw = request.headers.get(TRACEPARENT_HEADER)
    if traceparent_raw:
        parsed = parse_traceparent(traceparent_raw)
        if parsed is not None:
            return parsed
    request_id = request.headers.get(REQUEST_ID_HEADER)
    if request_id:
        return _trace_context_from_request_id(request_id)
    return new_trace_context()


_payload_logger = logging.getLogger("py_core.payload.http")


class TracingMiddleware(BaseHTTPMiddleware):
    async def dispatch(self, request: Request, call_next: Callable[[Request], Any]) -> Any:
        ctx = _resolve_trace_context_from_request(request)
        set_current_trace_context(ctx)
        try:
            return await call_next(request)
        finally:
            set_current_trace_context(None)


class PayloadLoggingMiddleware(BaseHTTPMiddleware):
    def __init__(self, app: Any, policy: RedactionPolicy) -> None:
        super().__init__(app)
        self._policy = policy

    async def dispatch(self, request: Request, call_next: Callable[[Request], Any]) -> Any:
        if not _payload_logger.isEnabledFor(logging.DEBUG):
            return await call_next(request)

        if request.method in ("POST", "PUT", "PATCH"):
            body_bytes = await request.body()
            if body_bytes:
                try:
                    parsed = json.loads(body_bytes)
                except (json.JSONDecodeError, TypeError):
                    _payload_logger.debug(
                        "http_request",
                        extra={
                            "method": request.method,
                            "path": str(request.url.path),
                            "body": "[non-json]",
                        },
                    )
                else:
                    if isinstance(parsed, dict):
                        redacted = self._policy.redact_dict(parsed)
                    else:
                        redacted = "[non-object-json]"
                    _payload_logger.debug(
                        "http_request",
                        extra={
                            "method": request.method,
                            "path": str(request.url.path),
                            "body": redacted,
                        },
                    )

        response = await call_next(request)
        _payload_logger.debug(
            "http_response",
            extra={
                "method": request.method,
                "path": str(request.url.path),
                "status_code": response.status_code,
            },
        )
        return response


def create_app(
    *,
    title: str,
    description: str,
    lifespan: Callable[[Any], AsyncGenerator[None, None]] | None = None,
    routers: list[APIRouter] | None = None,
    service_name: str | None = None,
    enable_request_tracing: bool = True,
    enable_payload_logging: bool | None = None,
    payload_redaction_policy: RedactionPolicy | None = None,
) -> FastAPI:
    """Create a FastAPI application with a standard /health endpoint.

    Args:
        title: OpenAPI title for the service.
        description: OpenAPI description.
        lifespan: Optional async context manager for startup/shutdown.
        routers: Additional APIRouters to mount.
        service_name: Optional service name (reserved for future use).
        enable_request_tracing: When True, install W3C trace context middleware.
        enable_payload_logging: When True, log redacted HTTP bodies at DEBUG. When
            None, ``LOG_PAYLOADS`` env var is consulted.
        payload_redaction_policy: Policy for redaction; defaults to ``DEFAULT_POLICY``.
    """
    _ = service_name
    resolved_payload_logging = _should_log_payloads(enable_payload_logging)
    effective_policy = payload_redaction_policy or DEFAULT_POLICY
    app = FastAPI(title=title, description=description, lifespan=lifespan)
    app.add_middleware(
        CORSMiddleware,
        allow_origins=["http://localhost:4200"],
        allow_methods=["*"],
        allow_headers=["*"],
    )
    if resolved_payload_logging:
        app.add_middleware(PayloadLoggingMiddleware, policy=effective_policy)
    if enable_request_tracing:
        app.add_middleware(TracingMiddleware)
    app.include_router(health_router)
    for router in routers or []:
        app.include_router(router)
    return app
