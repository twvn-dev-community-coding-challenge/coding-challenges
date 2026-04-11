"""FastAPI app factory with shared /health endpoint."""

from __future__ import annotations

from collections.abc import AsyncGenerator, Callable
from typing import Any

from fastapi import APIRouter, FastAPI

health_router = APIRouter()


@health_router.get("/health")
def health() -> dict[str, str]:
    return {"status": "ok"}


def create_app(
    *,
    title: str,
    description: str,
    lifespan: Callable[[Any], AsyncGenerator[None, None]] | None = None,
    routers: list[APIRouter] | None = None,
) -> FastAPI:
    """Create a FastAPI application with a standard /health endpoint.

    Args:
        title: OpenAPI title for the service.
        description: OpenAPI description.
        lifespan: Optional async context manager for startup/shutdown.
        routers: Additional APIRouters to mount.
    """
    app = FastAPI(title=title, description=description, lifespan=lifespan)
    app.include_router(health_router)
    for router in routers or []:
        app.include_router(router)
    return app
