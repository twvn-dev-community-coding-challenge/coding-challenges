from main import app


def test_openapi_schema_has_title_and_health_path() -> None:
    schema = app.openapi()
    assert schema["info"]["title"] == "Notification Service"
    assert "/health" in schema["paths"]
    assert "get" in schema["paths"]["/health"]
