"""Root logging configuration with trace context injection.

Console output uses human-readable format; file appender uses structured JSON.
"""

from __future__ import annotations

import logging
import os
import sys
from logging.handlers import RotatingFileHandler

from pythonjsonlogger.json import JsonFormatter

from py_core.tracing import get_current_trace_context

_HUMAN_LOG_FORMAT = (
    "%(asctime)s %(levelname)s [%(name)s] [trace=%(trace_id)s span=%(span_id)s] %(message)s"
)


class TraceContextFilter(logging.Filter):
    """Injects trace_id, span_id, and optional service name into each LogRecord."""

    def __init__(self, service_name: str | None = None) -> None:
        super().__init__()
        self._service_name = service_name

    def filter(self, record: logging.LogRecord) -> bool:
        ctx = get_current_trace_context()
        if ctx is None:
            record.trace_id = "-"
            record.span_id = "-"
        else:
            record.trace_id = ctx.trace_id
            record.span_id = ctx.span_id
        if self._service_name is not None:
            record.service = self._service_name
        return True


def _resolve_level(level_name: str) -> int:
    mapping = logging.getLevelNamesMapping()
    upper = level_name.upper()
    if upper not in mapping:
        msg = f"Unknown log level: {level_name!r}"
        raise ValueError(msg)
    return mapping[upper]


def _build_json_formatter() -> JsonFormatter:
    return JsonFormatter(
        fmt=("timestamp", "levelname", "name", "message", "trace_id", "span_id", "service"),
        rename_fields={"levelname": "level"},
        timestamp=True,
    )


def _should_log_payloads(explicit: bool | None) -> bool:
    if explicit is not None:
        return explicit
    return os.environ.get("LOG_PAYLOADS", "").lower() in ("1", "true", "yes")


def configure_logging(
    *,
    level: str | None = None,
    service_name: str | None = None,
    log_dir: str | None = None,
) -> None:
    """Configure root logging: human-readable to stderr, JSON to file.

    Reads ``LOG_LEVEL`` and optional ``LOG_DIR`` env vars when arguments are
    omitted. Installs a filter that adds ``trace_id``, ``span_id``, and
    optional ``service_name`` to every :class:`logging.LogRecord`.

    Console (stderr) always uses a human-readable format.
    File appender (when ``log_dir`` is set) always uses structured JSON so
    that payload ``extra`` fields are captured.
    """
    level_str = level if level is not None else os.environ.get("LOG_LEVEL", "INFO")

    resolved_level = _resolve_level(level_str)
    effective_log_dir = log_dir if log_dir is not None else os.environ.get("LOG_DIR")

    root = logging.getLogger()
    root.handlers.clear()
    root.setLevel(resolved_level)

    trace_filter = TraceContextFilter(service_name=service_name)

    stream_handler = logging.StreamHandler(stream=sys.stderr)
    stream_handler.setLevel(resolved_level)
    stream_handler.addFilter(trace_filter)
    stream_handler.setFormatter(logging.Formatter(fmt=_HUMAN_LOG_FORMAT))
    root.addHandler(stream_handler)

    if effective_log_dir:
        os.makedirs(effective_log_dir, exist_ok=True)
        log_filename = f"{service_name}.log" if service_name else "app.log"
        file_path = os.path.join(effective_log_dir, log_filename)
        file_handler = RotatingFileHandler(
            file_path,
            maxBytes=10_000_000,
            backupCount=5,
        )
        file_handler.setLevel(resolved_level)
        file_handler.addFilter(trace_filter)
        file_handler.setFormatter(_build_json_formatter())
        root.addHandler(file_handler)
