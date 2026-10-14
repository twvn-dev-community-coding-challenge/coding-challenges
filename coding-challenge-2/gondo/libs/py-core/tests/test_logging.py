"""Tests for py_core.logging configuration and trace context injection."""

from __future__ import annotations

import io
import json
import logging
from logging.handlers import RotatingFileHandler
from pathlib import Path

import pytest

from py_core.logging import configure_logging
from py_core.tracing import TraceContext, set_current_trace_context


@pytest.fixture(autouse=True)
def _reset_root_logging() -> None:
    root = logging.getLogger()
    root.handlers.clear()
    root.setLevel(logging.WARNING)
    yield
    root.handlers.clear()
    root.setLevel(logging.WARNING)


def test_configure_logging_sets_root_level() -> None:
    configure_logging(level="DEBUG")
    assert logging.getLogger().level == logging.DEBUG


def test_configure_logging_human_console_format() -> None:
    buf = io.StringIO()
    configure_logging(level="INFO")
    handler = logging.getLogger().handlers[0]
    handler.stream = buf

    logger = logging.getLogger("test.human")
    logger.info("hello-human")

    output = buf.getvalue()
    assert "hello-human" in output
    assert "INFO" in output
    assert "test.human" in output


def test_configure_logging_console_includes_trace_placeholders() -> None:
    buf = io.StringIO()
    configure_logging(level="INFO")
    handler = logging.getLogger().handlers[0]
    handler.stream = buf

    set_current_trace_context(None)
    logging.getLogger("test.dash").info("no-trace")

    output = buf.getvalue()
    assert "trace=-" in output
    assert "span=-" in output


def test_trace_context_filter_injects_trace_fields() -> None:
    buf = io.StringIO()
    configure_logging(level="INFO")
    handler = logging.getLogger().handlers[0]
    handler.stream = buf

    tid = "a" * 32
    sid = "b" * 16
    ctx = TraceContext(trace_id=tid, span_id=sid, parent_span_id=None)
    try:
        set_current_trace_context(ctx)
        logging.getLogger("test.trace").info("with-trace")
    finally:
        set_current_trace_context(None)

    output = buf.getvalue()
    assert tid in output
    assert sid in output


def test_configure_logging_reads_env_vars(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("LOG_LEVEL", "ERROR")

    buf = io.StringIO()
    configure_logging()
    handler = logging.getLogger().handlers[0]
    handler.stream = buf

    assert logging.getLogger().level == logging.ERROR

    logging.getLogger("env.test").warning("should-not-appear")
    assert buf.getvalue() == ""

    logging.getLogger("env.test").error("err-line")
    assert "err-line" in buf.getvalue()


def test_configure_logging_idempotent() -> None:
    configure_logging(level="INFO")
    configure_logging(level="DEBUG")
    assert len(logging.getLogger().handlers) == 1


def test_configure_logging_with_log_dir_creates_file(tmp_path: Path) -> None:
    log_dir = str(tmp_path / "logs")
    configure_logging(level="INFO", log_dir=log_dir, service_name="file-svc")
    logging.getLogger("test.file").info("file-hello-line")
    log_file = tmp_path / "logs" / "file-svc.log"
    assert log_file.is_file()
    content = log_file.read_text().strip()
    assert "file-hello-line" in content


def test_configure_logging_file_handler_uses_service_name(tmp_path: Path) -> None:
    log_dir = str(tmp_path / "out")
    configure_logging(level="INFO", log_dir=log_dir, service_name="my-svc")
    logging.getLogger("x").info("m")
    assert (tmp_path / "out" / "my-svc.log").is_file()


def test_configure_logging_file_handler_default_app_name(tmp_path: Path) -> None:
    log_dir = str(tmp_path / "out")
    configure_logging(level="INFO", log_dir=log_dir)
    logging.getLogger("y").info("n")
    assert (tmp_path / "out" / "app.log").is_file()


def test_configure_logging_file_uses_json_format(tmp_path: Path) -> None:
    log_dir = str(tmp_path / "logs")
    configure_logging(level="INFO", log_dir=log_dir, service_name="json-svc")

    logging.getLogger("test.json").info("hello-json")

    log_file = tmp_path / "logs" / "json-svc.log"
    data = json.loads(log_file.read_text().strip())
    assert data["message"] == "hello-json"
    assert "trace_id" in data
    assert "span_id" in data


def test_configure_logging_file_json_includes_service_name(tmp_path: Path) -> None:
    log_dir = str(tmp_path / "logs")
    configure_logging(level="INFO", log_dir=log_dir, service_name="payments")

    logging.getLogger("test.svc").info("svc-msg")

    log_file = tmp_path / "logs" / "payments.log"
    data = json.loads(log_file.read_text().strip())
    assert data.get("service") == "payments"


def test_configure_logging_file_json_has_trace_fields(tmp_path: Path) -> None:
    log_dir = str(tmp_path / "logs")
    configure_logging(level="INFO", log_dir=log_dir, service_name="trace-svc")

    tid = "f" * 32
    sid = "e" * 16
    ctx = TraceContext(trace_id=tid, span_id=sid, parent_span_id=None)
    try:
        set_current_trace_context(ctx)
        logging.getLogger("test.tracefile").info("tf-msg")
    finally:
        set_current_trace_context(None)

    log_file = tmp_path / "logs" / "trace-svc.log"
    data = json.loads(log_file.read_text().strip())
    assert data["message"] == "tf-msg"
    assert data["trace_id"] == tid


def test_configure_logging_file_and_console_both_active(tmp_path: Path) -> None:
    log_dir = str(tmp_path / "logs")
    configure_logging(level="INFO", log_dir=log_dir)
    root = logging.getLogger()
    assert len(root.handlers) == 2
    assert any(isinstance(h, logging.StreamHandler) for h in root.handlers)
    assert any(isinstance(h, RotatingFileHandler) for h in root.handlers)


def test_configure_logging_creates_log_dir_if_missing(tmp_path: Path) -> None:
    nested = tmp_path / "nested" / "deep" / "logs"
    assert not nested.exists()
    configure_logging(level="INFO", log_dir=str(nested))
    assert nested.is_dir()
