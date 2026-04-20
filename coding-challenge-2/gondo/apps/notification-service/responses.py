"""Shared HTTP JSON envelopes."""

from __future__ import annotations

from fastapi.responses import JSONResponse


def success_response(data: dict[str, object]) -> dict[str, object]:
    return {"data": data}


def error_response(
    code: str,
    message: str,
    status_code: int,
    details: dict[str, object] | None = None,
) -> JSONResponse:
    payload: dict[str, object] = {
        "error": {
            "code": code,
            "message": message,
            "details": details if details is not None else {},
        }
    }
    return JSONResponse(status_code=status_code, content=payload)
