from __future__ import annotations

import os

from generated import charging_pb2, charging_pb2_grpc
from py_core.app import create_app
from py_core.logging import _should_log_payloads, configure_logging
from py_core.server import grpc_lifespan

from grpc_server import ChargingGrpcServicer

_LOGS_DIR = os.path.join(os.path.dirname(__file__), "..", "..", "logs")
configure_logging(service_name="charging-service", log_dir=_LOGS_DIR)

_PAYLOAD_LOGGING = _should_log_payloads(None)

app = create_app(
    title="Charging Service",
    description="Cost estimation and recording",
    service_name="charging-service",
    enable_payload_logging=_PAYLOAD_LOGGING,
    lifespan=grpc_lifespan(
        servicers=[
            (
                charging_pb2_grpc.add_ChargingServiceServicer_to_server,
                ChargingGrpcServicer(),
            ),
        ],
        descriptors=[charging_pb2.DESCRIPTOR],
        bind="0.0.0.0:50052",
        enable_payload_logging=_PAYLOAD_LOGGING,
    ),
)
