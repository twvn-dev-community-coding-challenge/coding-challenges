"""HTTP API for the carrier bounded context registry (DDD — not on provider-service)."""

from __future__ import annotations

from fastapi import APIRouter, Query

from py_core.credentials_store import backend_name, secret_configured

from registry_loader import filter_carrier_entries

router = APIRouter(prefix="/registry", tags=["carrier-registry"])


@router.get("/carriers")
def list_carrier_registry(
    country_code: str = Query(..., min_length=2, max_length=2),
    carrier: str | None = Query(None, description="Optional MNO code filter"),
) -> dict[str, object]:
    """Return carrier (MNO) registry rows for a country; never includes raw secrets."""
    cc = country_code.strip().upper()
    entries = filter_carrier_entries(cc, carrier)
    return {
        "country_code": cc,
        "credentials_backend": backend_name(),
        "entries": [
            {
                "country_code": e.country_code,
                "carrier_code": e.carrier_code,
                "routing_hints": dict(e.routing_hints),
                "carrier_credentials_ref": e.carrier_credentials_ref or "",
                "carrier_credentials_configured": secret_configured(
                    e.carrier_credentials_ref,
                ),
            }
            for e in entries
        ],
    }
