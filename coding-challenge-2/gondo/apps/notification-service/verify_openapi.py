"""Fail if committed OpenAPI JSON does not match the live FastAPI app schema."""

from __future__ import annotations

import json
import sys
from pathlib import Path

from main import app


def _canonical(obj: object) -> str:
    """Stable string for comparing OpenAPI documents (key order normalized)."""
    return json.dumps(obj, sort_keys=True, separators=(",", ":"))


def verify() -> None:
    live = app.openapi()
    live_norm = _canonical(live)

    repo_root = Path(__file__).resolve().parent.parent.parent
    paths: list[tuple[Path, str]] = [
        (Path("openapi.json"), "apps/notification-service/openapi.json"),
        (
            repo_root / "docs" / "openapi" / "notification-service.openapi.json",
            "docs/openapi/notification-service.openapi.json",
        ),
    ]

    cwd = Path(__file__).resolve().parent
    errors: list[str] = []
    for path, label in paths:
        resolved = path if path.is_absolute() else cwd / path
        if not resolved.is_file():
            errors.append(f"Missing {label} at {resolved}")
            continue
        committed = json.loads(resolved.read_text(encoding="utf-8"))
        if _canonical(committed) != live_norm:
            errors.append(
                f"Drift: {label} does not match live FastAPI schema. "
                f"Run: yarn nx run notification-service:generate-openapi",
            )

    if errors:
        print("OpenAPI verification failed:", file=sys.stderr)
        for e in errors:
            print(f"  - {e}", file=sys.stderr)
        raise SystemExit(1)

    print("OpenAPI artifacts match live app (openapi.json + docs/openapi export).")


if __name__ == "__main__":
    try:
        verify()
    except Exception as err:
        print(f"verify_openapi failed: {err}", file=sys.stderr)
        raise SystemExit(1) from err
