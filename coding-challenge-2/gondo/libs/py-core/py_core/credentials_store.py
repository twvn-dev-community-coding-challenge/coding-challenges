"""Shared credential resolution (mock JSON or AWS Secrets Manager / Floci).

Used by provider-service and carrier-service registries. Never log raw secret values.
"""

from __future__ import annotations

import json
import logging
import os
from functools import lru_cache
from pathlib import Path
from typing import Any

logger = logging.getLogger(__name__)


def backend_name() -> str:
    return os.environ.get("CREDENTIALS_BACKEND", "mock").strip().lower()


def _gondo_root() -> Path:
    # py_core/credentials_store.py -> gondo/
    return Path(__file__).resolve().parent.parent.parent.parent


def default_mock_credentials_path() -> Path:
    override = os.environ.get("MOCK_CREDENTIALS_PATH")
    if override:
        return Path(override).expanduser().resolve()
    return _gondo_root() / "infra" / "credentials" / "mock" / "secrets.json"


@lru_cache(maxsize=1)
def _load_mock_store() -> dict[str, Any]:
    path = default_mock_credentials_path()
    if not path.is_file():
        logger.warning("mock_credentials_file_missing", extra={"path": str(path)})
        return {}
    try:
        raw = json.loads(path.read_text(encoding="utf-8"))
    except (json.JSONDecodeError, OSError) as exc:
        logger.warning("mock_credentials_unreadable", extra={"path": str(path), "error": str(exc)})
        return {}
    if not isinstance(raw, dict):
        return {}
    return raw


def clear_credentials_cache() -> None:
    _load_mock_store.cache_clear()


def _mock_configured(secret_id: str) -> bool:
    data = _load_mock_store().get(secret_id)
    if data is None:
        return False
    if isinstance(data, dict) and data:
        return any(bool(v) for v in data.values())
    return bool(data)


def _aws_secret_configured(secret_id: str) -> bool:
    try:
        import boto3
        from botocore.exceptions import ClientError
    except ImportError:
        logger.warning("boto3_missing_for_aws_credentials")
        return False
    endpoint = os.environ.get("AWS_ENDPOINT_URL")
    region = os.environ.get("AWS_REGION", "us-east-1")
    client = boto3.client(
        "secretsmanager",
        region_name=region,
        endpoint_url=endpoint or None,
        aws_access_key_id=os.environ.get("AWS_ACCESS_KEY_ID", "test"),
        aws_secret_access_key=os.environ.get("AWS_SECRET_ACCESS_KEY", "test"),
    )
    try:
        client.get_secret_value(SecretId=secret_id)
    except ClientError as exc:
        code = exc.response.get("Error", {}).get("Code", "")
        if code in ("ResourceNotFoundException", "SecretsManager.ResourceNotFoundException"):
            return False
        logger.warning("aws_secret_lookup_failed", extra={"secret_id": secret_id, "code": code})
        return False
    except OSError as exc:
        logger.warning("aws_secret_lookup_error", extra={"secret_id": secret_id, "error": str(exc)})
        return False
    return True


def secret_configured(secret_id: str | None) -> bool:
    """Return True if the secret id resolves in the active backend."""
    if not secret_id or not str(secret_id).strip():
        return False
    sid = str(secret_id).strip()
    b = backend_name()
    if b in ("mock", "", "file"):
        return _mock_configured(sid)
    if b in ("aws", "aws_secrets_manager", "secretsmanager"):
        return _aws_secret_configured(sid)
    logger.warning("unknown_credentials_backend", extra={"backend": b})
    return False
