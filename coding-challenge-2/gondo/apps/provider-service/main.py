from __future__ import annotations

import os

from py_core.app import create_app
from py_core.logging import _should_log_payloads, configure_logging

from lifespan import build_provider_lifespan

_LOGS_DIR = os.path.join(os.path.dirname(__file__), "..", "..", "logs")
configure_logging(service_name="provider-service", log_dir=_LOGS_DIR)

_PAYLOAD_LOGGING = _should_log_payloads(None)

app = create_app(
    title="Provider Service",
    description="Provider registry, routing rules, and CQRS dispatch events",
    service_name="provider-service",
    enable_payload_logging=_PAYLOAD_LOGGING,
    lifespan=build_provider_lifespan(enable_payload_logging=_PAYLOAD_LOGGING),
)
