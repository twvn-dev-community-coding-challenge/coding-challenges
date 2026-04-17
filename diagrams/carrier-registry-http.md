# Carrier registry HTTP (bounded context)

- **Not** **`infra/provider-registry`** — carrier MNO config lives in **`infra/carrier-registry/`**.
- **Loader / env diagram:** **[registry-config-loading.md](registry-config-loading.md)**.

**carrier-service** exposes read-only metadata at **`GET /registry/carriers`** (no raw secrets).

```mermaid
sequenceDiagram
    autonumber
    actor Op as Operator / script
    participant CS as carrier-service :8004
    participant CR as infra/carrier-registry
    participant Cred as credentials_store<br/>mock / AWS

    Op->>CS: GET /registry/carriers?country_code=VN&carrier=…
    CS->>CR: registry_loader — CARRIER_REGISTRY_ROOT
    CS->>Cred: secret_configured(ref)
    CS-->>Op: country_code, entries[], credentials_backend
```

Response shape (conceptual): **`entries[]`** with **`carrier_code`**, **`routing_hints`**, **`carrier_credentials_ref`**, **`carrier_credentials_configured`**.

## Code references

- `apps/carrier-service/registry_routes.py`
- `apps/carrier-service/registry_loader.py`
- `libs/py-core/py_core/credentials_store.py`
