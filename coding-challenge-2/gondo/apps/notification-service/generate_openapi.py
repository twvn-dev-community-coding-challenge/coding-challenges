"""Generate the OpenAPI JSON contract from the FastAPI app."""

import json
import sys
from pathlib import Path

from main import app


def generate() -> None:
    schema = app.openapi()
    compact = json.dumps(schema)
    service_path = Path("openapi.json")
    service_path.write_text(json.dumps(schema, indent=2) + "\n")
    print(f"OpenAPI contract written to {service_path} ({len(compact)} bytes)")

    repo_root = Path(__file__).resolve().parent.parent.parent
    docs_path = repo_root / "docs" / "openapi" / "notification-service.openapi.json"
    docs_path.parent.mkdir(parents=True, exist_ok=True)
    docs_path.write_text(json.dumps(schema, indent=2) + "\n")
    print(f"Platform SMS API export: {docs_path}")


if __name__ == "__main__":
    try:
        generate()
    except Exception as err:
        print(f"Failed to generate OpenAPI contract: {err}", file=sys.stderr)
        raise SystemExit(1) from err
