"""gRPC server factory with automatic reflection registration."""

from __future__ import annotations

import logging
from collections.abc import AsyncGenerator, Callable, Sequence
from contextlib import asynccontextmanager
from typing import Any

from google.protobuf import descriptor as _descriptor
from google.protobuf.message import Message
from grpc import aio
from grpc import _utilities
from grpc_reflection.v1alpha import reflection

from py_core.logging import _should_log_payloads
from py_core.redact import DEFAULT_POLICY, RedactionPolicy
from py_core.tracing import (
    TraceContext,
    metadata_to_trace_context,
    new_trace_context,
    set_current_trace_context,
)


def _invocation_metadata_pairs(
    invocation_metadata: object | None,
) -> tuple[tuple[str, str], ...]:
    if not invocation_metadata:
        return ()
    pairs: list[tuple[str, str]] = []
    for item in invocation_metadata:
        key = getattr(item, "key", None)
        val = getattr(item, "value", None)
        if key is None and isinstance(item, (tuple, list)) and len(item) >= 2:
            key, val = item[0], item[1]
        if key is None:
            continue
        ks = key.decode("utf-8") if isinstance(key, bytes) else str(key)
        vs = val.decode("utf-8") if isinstance(val, bytes) else str(val)
        pairs.append((ks, vs))
    return tuple(pairs)


_grpc_payload_logger = logging.getLogger("py_core.payload.grpc")


def _wrap_rpc_method_handler(
    ctx: TraceContext | None,
    rmh: Any,
    *,
    method: str,
    enable_payload_logging: bool,
    policy: RedactionPolicy,
) -> Any:
    uu, us, su, ss = rmh.unary_unary, rmh.unary_stream, rmh.stream_unary, rmh.stream_stream

    async def wrap_uu(request: object, context: object) -> object:
        if ctx is not None:
            set_current_trace_context(ctx)
        try:
            if (
                enable_payload_logging
                and _grpc_payload_logger.isEnabledFor(logging.DEBUG)
                and isinstance(request, Message)
            ):
                redacted = policy.redact_proto(request)
                _grpc_payload_logger.debug(
                    "grpc_server_request",
                    extra={"method": method, "body": redacted},
                )
            result = await uu(request, context)
            if (
                enable_payload_logging
                and _grpc_payload_logger.isEnabledFor(logging.DEBUG)
                and isinstance(result, Message)
            ):
                redacted_resp = policy.redact_proto(result)
                _grpc_payload_logger.debug(
                    "grpc_server_response",
                    extra={"method": method, "body": redacted_resp},
                )
            return result
        finally:
            if ctx is not None:
                set_current_trace_context(None)

    async def wrap_us(request: object, context: object):
        if ctx is not None:
            set_current_trace_context(ctx)
        try:
            async for item in us(request, context):
                yield item
        finally:
            if ctx is not None:
                set_current_trace_context(None)

    async def wrap_su(request_iterator: object, context: object) -> object:
        if ctx is not None:
            set_current_trace_context(ctx)
        try:
            return await su(request_iterator, context)
        finally:
            if ctx is not None:
                set_current_trace_context(None)

    async def wrap_ss(request_iterator: object, context: object):
        if ctx is not None:
            set_current_trace_context(ctx)
        try:
            async for item in ss(request_iterator, context):
                yield item
        finally:
            if ctx is not None:
                set_current_trace_context(None)

    return _utilities.RpcMethodHandler(
        rmh.request_streaming,
        rmh.response_streaming,
        rmh.request_deserializer,
        rmh.response_serializer,
        wrap_uu if uu is not None else None,
        wrap_us if us is not None else None,
        wrap_su if su is not None else None,
        wrap_ss if ss is not None else None,
    )


class TracingServerInterceptor(aio.ServerInterceptor):
    def __init__(
        self,
        *,
        enable_tracing: bool = True,
        enable_payload_logging: bool = False,
        payload_policy: RedactionPolicy | None = None,
    ) -> None:
        self._enable_tracing = enable_tracing
        self._enable_payload_logging = enable_payload_logging
        self._payload_policy = payload_policy or DEFAULT_POLICY

    async def intercept_service(
        self,
        continuation: Callable[[Any], Any],
        handler_call_details: Any,
    ) -> Any:
        if self._enable_tracing:
            pairs = _invocation_metadata_pairs(handler_call_details.invocation_metadata)
            ctx = metadata_to_trace_context(pairs)
            if ctx is None:
                ctx = new_trace_context()
        else:
            ctx = None
        rmh = await continuation(handler_call_details)
        if rmh is None:
            return None
        method = handler_call_details.method
        if isinstance(method, bytes):
            method = method.decode("utf-8")
        else:
            method = str(method)
        return _wrap_rpc_method_handler(
            ctx,
            rmh,
            method=method,
            enable_payload_logging=self._enable_payload_logging,
            policy=self._payload_policy,
        )


async def create_grpc_server(
    *,
    servicers: Sequence[tuple[Callable[..., None], Any]],
    descriptors: Sequence[_descriptor.FileDescriptor],
    bind: str,
    enable_tracing: bool = True,
    enable_payload_logging: bool | None = None,
    payload_redaction_policy: RedactionPolicy | None = None,
) -> aio.Server:
    """Create, configure, and start an insecure gRPC server.

    Args:
        servicers: Pairs of ``(add_*Servicer_to_server, servicer_instance)``.
        descriptors: Proto ``DESCRIPTOR`` objects for reflection discovery.
        bind: Address to bind (e.g. ``"0.0.0.0:50051"``).
        enable_tracing: When True, install trace context server interceptors.
        enable_payload_logging: When True, log redacted unary payloads at DEBUG.
            When None, ``LOG_PAYLOADS`` env var is consulted.
        payload_redaction_policy: Policy for redaction; defaults to ``DEFAULT_POLICY``.
    """
    resolved_payload = _should_log_payloads(enable_payload_logging)
    effective_policy = payload_redaction_policy or DEFAULT_POLICY
    interceptors: tuple[aio.ServerInterceptor, ...] | None = None
    if enable_tracing or resolved_payload:
        interceptors = (
            TracingServerInterceptor(
                enable_tracing=enable_tracing,
                enable_payload_logging=resolved_payload,
                payload_policy=effective_policy,
            ),
        )
    server = aio.server(interceptors=interceptors)
    for add_fn, servicer in servicers:
        add_fn(servicer, server)
    service_names = [svc.full_name for fd in descriptors for svc in fd.services_by_name.values()]
    reflection.enable_server_reflection(
        [*service_names, reflection.SERVICE_NAME],
        server,
    )
    server.add_insecure_port(bind)
    await server.start()
    return server


def grpc_lifespan(
    *,
    servicers: Sequence[tuple[Callable[..., None], Any]],
    descriptors: Sequence[_descriptor.FileDescriptor],
    bind: str,
    grace: float = 5.0,
    enable_tracing: bool = True,
    enable_payload_logging: bool | None = None,
    payload_redaction_policy: RedactionPolicy | None = None,
) -> Callable[[Any], AsyncGenerator[None, None]]:
    """Return a FastAPI lifespan that manages a gRPC server.

    The gRPC server starts on app startup and stops gracefully on shutdown.
    """

    @asynccontextmanager
    async def _lifespan(_app: Any) -> AsyncGenerator[None, None]:
        server = await create_grpc_server(
            servicers=servicers,
            descriptors=descriptors,
            bind=bind,
            enable_tracing=enable_tracing,
            enable_payload_logging=enable_payload_logging,
            payload_redaction_policy=payload_redaction_policy,
        )
        yield
        await server.stop(grace)

    return _lifespan
