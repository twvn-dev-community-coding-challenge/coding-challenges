"""Generate the OpenAPI JSON contract from the FastAPI app."""

import json
import sys
from pathlib import Path

from main import app


def generate() -> None:
    schema = app.openapi()
    output_path = Path("openapi.json")
    output_path.write_text(json.dumps(schema, indent=2) + "\n")
    print(f"OpenAPI contract written to {output_path} ({len(json.dumps(schema))} bytes)")


if __name__ == "__main__":
    try:
        generate()
    except Exception as err:
        print(f"Failed to generate OpenAPI contract: {err}", file=sys.stderr)
        raise SystemExit(1) from err
