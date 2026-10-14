"""Carrier service: consumes dispatch events and calls provider HTTP APIs (CQRS handlers)."""

from __future__ import annotations

import os

from py_core.app import create_app
from py_core.logging import _should_log_payloads, configure_logging

from lifespan import carrier_lifespan
from registry_routes import router as carrier_registry_router

_LOGS_DIR = os.path.join(os.path.dirname(__file__), "..", "..", "logs")
configure_logging(service_name="carrier-service", log_dir=_LOGS_DIR)

_PAYLOAD_LOGGING = _should_log_payloads(None)

app = create_app(
    title="Carrier Service",
    description="SMS dispatch execution, outcome events, and carrier-registry (MNO bounded context)",
    service_name="carrier-service",
    enable_payload_logging=_PAYLOAD_LOGGING,
    lifespan=carrier_lifespan,
    routers=[carrier_registry_router],
)
