"""gRPC client for notification-service -> charging-service."""

from __future__ import annotations

import os
import uuid

from datetime import datetime, timezone

from google.protobuf import timestamp_pb2

from generated import charging_pb2, charging_pb2_grpc
from py_core.client import insecure_channel
from py_core.logging import _should_log_payloads
from py_core.proto_utils import as_utc, datetime_to_timestamp

CHARGING_GRPC_TARGET = os.environ.get("CHARGING_GRPC_TARGET", "localhost:50052")
_PAYLOAD_LOGGING = _should_log_payloads(None)


async def estimate_cost_grpc(
    *,
    message_id: str,
    provider_id: str,
    country_code: str,
    carrier: str,
    as_of: timestamp_pb2.Timestamp,
) -> charging_pb2.EstimateCostResponse:
    async with insecure_channel(
        CHARGING_GRPC_TARGET, enable_payload_logging=_PAYLOAD_LOGGING
    ) as channel:
        stub = charging_pb2_grpc.ChargingServiceStub(channel)
        request = charging_pb2.EstimateCostRequest(
            message_id=message_id,
            provider_id=provider_id,
            country_code=country_code.strip().upper(),
            carrier=carrier,
            as_of=as_of,
        )
        return await stub.EstimateCost(request)


async def record_actual_cost_grpc(
    *,
    message_id: str,
    provider_id: str,
    actual_cost: float,
    currency: str,
    callback_state: str,
    idempotency_key: str,
    recorded_at: timestamp_pb2.Timestamp | None = None,
    provider_event_id: str | None = None,
) -> charging_pb2.RecordActualCostResponse:
    ts = recorded_at or datetime_to_timestamp(as_utc(datetime.now(timezone.utc)))
    ev = provider_event_id or str(uuid.uuid4())
    async with insecure_channel(
        CHARGING_GRPC_TARGET, enable_payload_logging=_PAYLOAD_LOGGING
    ) as channel:
        stub = charging_pb2_grpc.ChargingServiceStub(channel)
        request = charging_pb2.RecordActualCostRequest(
            message_id=message_id,
            provider_id=provider_id,
            provider_event_id=ev,
            idempotency_key=idempotency_key,
            actual_cost=actual_cost,
            currency=currency,
            callback_state=callback_state,
            recorded_at=ts,
        )
        return await stub.RecordActualCost(request)
