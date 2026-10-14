from main import app
from verify_openapi import verify


def test_openapi_schema_has_title_and_health_path() -> None:
    schema = app.openapi()
    assert schema["info"]["title"] == "Notification Service"
    assert "/health" in schema["paths"]
    assert "get" in schema["paths"]["/health"]


def test_committed_openapi_json_matches_live_app() -> None:
    """Regenerate with `generate_openapi.py` if this fails (drift)."""
    verify()
