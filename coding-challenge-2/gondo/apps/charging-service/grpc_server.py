"""gRPC servicer for charging-service."""

from __future__ import annotations

import logging

import grpc
from grpc import aio

from generated import charging_pb2, charging_pb2_grpc
from py_core.proto_utils import as_utc, datetime_to_timestamp
from rates import (
    ActualCostRecord,
    EstimateRecord,
    estimate_cost,
    estimate_cost_batch,
    record_actual_cost,
)

logger = logging.getLogger(__name__)


def _estimate_record_to_proto(rec: EstimateRecord) -> charging_pb2.EstimateCostResponse:
    return charging_pb2.EstimateCostResponse(
        estimate_id=rec.estimate_id,
        estimated_cost=rec.estimated_cost,
        currency=rec.currency,
        rate_id=rec.rate_id,
        rate_version=rec.rate_version,
        created_at=datetime_to_timestamp(rec.created_at),
    )


def _batch_item_from_record(rec: EstimateRecord) -> charging_pb2.EstimateCostBatchItem:
    return charging_pb2.EstimateCostBatchItem(
        provider_id=rec.provider_id,
        estimate_id=rec.estimate_id,
        estimated_cost=rec.estimated_cost,
        currency=rec.currency,
        rate_id=rec.rate_id,
        rate_version=rec.rate_version,
    )


def _actual_record_to_proto(
    rec: ActualCostRecord,
) -> charging_pb2.RecordActualCostResponse:
    return charging_pb2.RecordActualCostResponse(
        actual_cost_id=rec.actual_cost_id,
        message_id=rec.message_id,
        actual_cost=rec.actual_cost,
        idempotent_replay=rec.idempotent_replay,
    )


class ChargingGrpcServicer(charging_pb2_grpc.ChargingServiceServicer):
    async def EstimateCost(
        self,
        request: charging_pb2.EstimateCostRequest,
        context: aio.ServicerContext,
    ) -> charging_pb2.EstimateCostResponse:
        if not request.HasField("as_of"):
            await context.abort(
                grpc.StatusCode.INVALID_ARGUMENT,
                "as_of is required",
            )
        as_of = as_utc(request.as_of.ToDatetime())
        logger.info(
            "charging_grpc_request",
            extra={
                "grpc_method": "EstimateCost",
                "message_id": request.message_id,
                "provider_id": request.provider_id,
                "country_code": request.country_code,
                "carrier": request.carrier,
            },
        )
        rec = await estimate_cost(
            request.message_id,
            request.provider_id,
            request.country_code,
            request.carrier,
            as_of,
        )
        if rec is None:
            await context.abort(
                grpc.StatusCode.NOT_FOUND,
                "RATE_NOT_AVAILABLE",
            )
        logger.info(
            "charging_grpc_response",
            extra={
                "grpc_method": "EstimateCost",
                "message_id": request.message_id,
                "estimate_id": rec.estimate_id,
                "estimated_cost": rec.estimated_cost,
                "currency": rec.currency,
            },
        )
        return _estimate_record_to_proto(rec)

    async def EstimateCostBatch(
        self,
        request: charging_pb2.EstimateCostBatchRequest,
        context: aio.ServicerContext,
    ) -> charging_pb2.EstimateCostBatchResponse:
        if not request.HasField("as_of"):
            await context.abort(
                grpc.StatusCode.INVALID_ARGUMENT,
                "as_of is required",
            )
        as_of = as_utc(request.as_of.ToDatetime())
        rows = await estimate_cost_batch(
            list(request.provider_ids),
            request.country_code,
            request.carrier,
            as_of,
        )
        return charging_pb2.EstimateCostBatchResponse(
            estimates=[_batch_item_from_record(r) for r in rows],
        )

    async def RecordActualCost(
        self,
        request: charging_pb2.RecordActualCostRequest,
        context: aio.ServicerContext,
    ) -> charging_pb2.RecordActualCostResponse:
        if not request.HasField("recorded_at"):
            await context.abort(
                grpc.StatusCode.INVALID_ARGUMENT,
                "recorded_at is required",
            )
        recorded_at = as_utc(request.recorded_at.ToDatetime())
        logger.info(
            "charging_grpc_request",
            extra={
                "grpc_method": "RecordActualCost",
                "message_id": request.message_id,
                "provider_id": request.provider_id,
                "callback_state": request.callback_state,
                "idempotency_key": request.idempotency_key,
            },
        )
        rec = record_actual_cost(
            message_id=request.message_id,
            provider_id=request.provider_id,
            provider_event_id=request.provider_event_id,
            idempotency_key=request.idempotency_key,
            actual_cost=request.actual_cost,
            currency=request.currency,
            callback_state=request.callback_state,
            recorded_at=recorded_at,
        )
        logger.info(
            "charging_grpc_response",
            extra={
                "grpc_method": "RecordActualCost",
                "message_id": request.message_id,
                "actual_cost_id": rec.actual_cost_id,
                "idempotent_replay": rec.idempotent_replay,
            },
        )
        return _actual_record_to_proto(rec)
