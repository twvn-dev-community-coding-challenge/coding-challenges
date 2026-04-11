"""gRPC client primitives (async channel helpers)."""

from __future__ import annotations

import logging
from collections.abc import AsyncIterator, Callable
from contextlib import asynccontextmanager

from google.protobuf.message import Message
from grpc import aio
from grpc.aio import Channel, ClientCallDetails, Metadata, UnaryUnaryClientInterceptor

from py_core.logging import _should_log_payloads
from py_core.redact import DEFAULT_POLICY, RedactionPolicy
from py_core.tracing import (
    get_current_trace_context,
    trace_context_to_metadata,
)


def _merge_client_metadata(
    existing: Metadata | None,
    extra: tuple[tuple[str, str], ...],
) -> Metadata:
    if existing is None:
        return Metadata(*extra)
    return Metadata(*(tuple(existing) + extra))


class TracingClientInterceptor(UnaryUnaryClientInterceptor):
    async def intercept_unary_unary(
        self,
        continuation: Callable[[ClientCallDetails, object], object],
        client_call_details: ClientCallDetails,
        request: object,
    ) -> object:
        ctx = get_current_trace_context()
        if ctx is None:
            return await continuation(client_call_details, request)
        md = trace_context_to_metadata(ctx)
        merged = _merge_client_metadata(client_call_details.metadata, md)
        new_details = client_call_details._replace(metadata=merged)
        return await continuation(new_details, request)


_client_payload_logger = logging.getLogger("py_core.payload.grpc")


class PayloadLoggingClientInterceptor(UnaryUnaryClientInterceptor):
    def __init__(self, policy: RedactionPolicy) -> None:
        self._policy = policy

    async def intercept_unary_unary(
        self,
        continuation: Callable[[ClientCallDetails, object], object],
        client_call_details: ClientCallDetails,
        request: object,
    ) -> object:
        if _client_payload_logger.isEnabledFor(logging.DEBUG):
            try:
                if isinstance(request, Message):
                    redacted = self._policy.redact_proto(request)
                    method = client_call_details.method
                    if isinstance(method, bytes):
                        method = method.decode("utf-8")
                    _client_payload_logger.debug(
                        "grpc_client_request",
                        extra={"method": method, "body": redacted},
                    )
            except Exception:
                _client_payload_logger.debug(
                    "grpc_client_payload_redact_failed",
                    exc_info=True,
                )
        return await continuation(client_call_details, request)


@asynccontextmanager
async def insecure_channel(
    target: str,
    *,
    enable_tracing: bool = True,
    enable_payload_logging: bool | None = None,
    payload_redaction_policy: RedactionPolicy | None = None,
) -> AsyncIterator[Channel]:
    """Open an insecure async gRPC channel to ``target`` (``host:port``).

    Yields a :class:`grpc.aio.Channel` for use with generated stubs.
    The channel is closed when the context exits (success or failure).
    """
    resolved_payload = _should_log_payloads(enable_payload_logging)
    effective_policy = payload_redaction_policy or DEFAULT_POLICY
    interceptors: list[UnaryUnaryClientInterceptor] = []
    if resolved_payload:
        interceptors.append(PayloadLoggingClientInterceptor(effective_policy))
    if enable_tracing:
        interceptors.append(TracingClientInterceptor())
    channel = aio.insecure_channel(
        target,
        interceptors=tuple(interceptors) if interceptors else None,
    )
    try:
        yield channel
    finally:
        await channel.close()
