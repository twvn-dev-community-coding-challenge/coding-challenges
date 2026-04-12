"""Provider gRPC servicer tests."""

from __future__ import annotations

import asyncio
from datetime import datetime, timezone
from unittest.mock import AsyncMock, patch

from google.protobuf import timestamp_pb2

from generated import provider_pb2
from grpc_server import ProviderGrpcServicer
from registry import RoutingCandidate as DomainRoutingCandidate


class FakeServicerContext:
    def __init__(self) -> None:
        self._code: object | None = None
        self._details: str | None = None

    async def abort(self, code: object, details: str) -> None:
        self._code = code
        self._details = details
        msg = f"gRPC abort: {code} {details}"
        raise RuntimeError(msg)


def test_provider_grpc_servicer_instantiates() -> None:
    servicer = ProviderGrpcServicer()
    assert servicer is not None


def _make_candidate(
    *,
    provider_id: str,
    provider_code: str,
    routing_rule_id: str,
    precedence: int,
) -> DomainRoutingCandidate:
    eff = datetime(2026, 1, 1, 0, 0, 0, tzinfo=timezone.utc)
    resolved = datetime(2026, 6, 15, 12, 0, 0, tzinfo=timezone.utc)
    return DomainRoutingCandidate(
        provider_id=provider_id,
        provider_code=provider_code,
        routing_rule_id=routing_rule_id,
        routing_rule_version=1,
        effective_from=eff,
        effective_to=None,
        resolved_at=resolved,
        precedence=precedence,
    )


async def _resolve_vn_viettel() -> provider_pb2.ResolveRoutingResponse:
    candidates = [
        _make_candidate(
            provider_id="prv_02",
            provider_code="VONAGE",
            routing_rule_id="rr_01",
            precedence=100,
        ),
        _make_candidate(
            provider_id="prv_01",
            provider_code="TWILIO",
            routing_rule_id="rr_02",
            precedence=90,
        ),
    ]
    with patch("grpc_server.resolve_routing", AsyncMock(return_value=candidates)):
        servicer = ProviderGrpcServicer()
        ts = timestamp_pb2.Timestamp()
        ts.FromDatetime(datetime(2026, 6, 15, 12, 0, 0, tzinfo=timezone.utc))
        req = provider_pb2.ResolveRoutingRequest(
            country_code="VN",
            carrier="VIETTEL",
            as_of=ts,
        )
        ctx = FakeServicerContext()
        return await servicer.ResolveRouting(req, ctx)


def test_resolve_routing_returns_candidates() -> None:
    resp = asyncio.run(_resolve_vn_viettel())
    assert len(resp.candidates) == 2
    ids = {c.routing_rule_id for c in resp.candidates}
    assert ids == {"rr_01", "rr_02"}
