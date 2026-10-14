"""Tests for shared mock/AWS credential presence checks."""

from __future__ import annotations

import json

import pytest

from py_core.credentials_store import (
    backend_name,
    clear_credentials_cache,
    secret_configured,
)


@pytest.fixture(autouse=True)
def _clear_creds_cache() -> None:
    clear_credentials_cache()
    yield
    clear_credentials_cache()


def test_secret_configured_mock_backend(
    monkeypatch: pytest.MonkeyPatch,
    tmp_path,
) -> None:
    path = tmp_path / "secrets.json"
    path.write_text(
        json.dumps({"carrier/VN/X": {"k": "v"}}),
        encoding="utf-8",
    )
    monkeypatch.setenv("CREDENTIALS_BACKEND", "mock")
    monkeypatch.setenv("MOCK_CREDENTIALS_PATH", str(path))
    clear_credentials_cache()
    assert backend_name() == "mock"
    assert secret_configured("carrier/VN/X") is True
    assert secret_configured("missing/id") is False
    assert secret_configured("") is False
    assert secret_configured(None) is False
