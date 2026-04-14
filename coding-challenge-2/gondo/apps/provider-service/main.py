from __future__ import annotations

import os

from generated import provider_pb2, provider_pb2_grpc
from grpc_server import ProviderGrpcServicer
from py_core.app import create_app
from py_core.logging import _should_log_payloads, configure_logging
from py_core.server import grpc_lifespan

_LOGS_DIR = os.path.join(os.path.dirname(__file__), "..", "..", "logs")
configure_logging(service_name="provider-service", log_dir=_LOGS_DIR)

_PAYLOAD_LOGGING = _should_log_payloads(None)

app = create_app(
    title="Provider Service",
    description="Provider registry and routing rules",
    service_name="provider-service",
    enable_payload_logging=_PAYLOAD_LOGGING,
    lifespan=grpc_lifespan(
        servicers=[
            (
                provider_pb2_grpc.add_ProviderServiceServicer_to_server,
                ProviderGrpcServicer(),
            ),
        ],
        descriptors=[provider_pb2.DESCRIPTOR],
        bind="0.0.0.0:50051",
        enable_payload_logging=_PAYLOAD_LOGGING,
    ),
)
