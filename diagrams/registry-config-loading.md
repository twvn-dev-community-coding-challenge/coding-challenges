# Registry config: provider-service vs carrier-service

On-disk YAML under **`coding-challenge-2/gondo/infra/`**. Override roots with **`PROVIDER_REGISTRY_ROOT`** / **`CARRIER_REGISTRY_ROOT`** (Docker: **`/app/infra/...`**).

**Credentials:** YAML holds **refs** only; **`libs/py-core/py_core/credentials_store.py`** resolves **mock JSON** or **AWS Secrets Manager** — never return raw secrets on APIs.

## Provider — `infra/provider-registry/`

Routing, merged **`provider_id`** across countries, **`api_endpoint`** for **`sms.dispatch.requested`**, **`credentials_ref`** for gRPC registry metadata.

```mermaid
flowchart TB
  subgraph Env["Environment"]
    PR["PROVIDER_REGISTRY_ROOT"]
  end

  subgraph PS["provider-service"]
    YR["yaml_registry.py<br/>load_all_provider_configs"]
    REG["registry.py + gRPC"]
    PD["cqrs/publish_dispatch.py"]
    GRPC["GetProviderRegistry"]
    OS["cqrs/outcome_subscriber.py<br/>logs outcomes only"]
  end

  YML[("infra/provider-registry")]

  PR --> YR
  YML --> YR
  YR --> REG
  YR --> PD
  YR --> GRPC
```

| Consumer | Reads |
|----------|--------|
| Routing / **SelectProvider** | PostgreSQL rules + **`yaml_registry`** merged rows |
| **`sms.dispatch.requested` JSON** | **`api_endpoint`** from selected provider row (`publish_dispatch`) |
| **GetProviderRegistry** | Filtered configs + **`credentials_configured`** |

---

## Carrier — `infra/carrier-registry/`

**Only carrier-service** loads this tree (DDD). Used for **`GET /registry/carriers`** and **`secret_configured`** on responses.

**Outbound dispatch HTTP** uses **`api_endpoint`** from the **NATS event payload** (set when provider published **`sms.dispatch.requested`**), not a second registry lookup in **`dispatch_handler`**.

```mermaid
flowchart TB
  subgraph EnvC["Environment"]
    CR["CARRIER_REGISTRY_ROOT"]
  end

  subgraph CS["carrier-service"]
    RL["registry_loader.py"]
    RR["registry_routes.py"]
    DH["cqrs/dispatch_handler.py<br/>HTTP uses event.api_endpoint"]
  end

  YMC[("infra/carrier-registry")]

  CR --> RL
  YMC --> RL
  RL --> RR
```

---

## Shared credentials

```mermaid
flowchart LR
  PY["provider YAML credentials_ref"]
  CY["carrier YAML carrier_credentials_ref"]
  CS2["py_core.credentials_store"]
  MOCK[("mock secrets.json / AWS SM")]

  PY --> CS2
  CY --> CS2
  CS2 --> MOCK
```

| Env (typical) | Role |
|---------------|------|
| `CREDENTIALS_BACKEND` | `mock` or `aws_secrets_manager` |
| `MOCK_CREDENTIALS_PATH` | Path to JSON for local dev |

## Code references

| Area | Path |
|------|------|
| Provider load / merge | `apps/provider-service/yaml_registry.py` |
| Publish payload | `apps/provider-service/cqrs/publish_dispatch.py` |
| Carrier load | `apps/carrier-service/registry_loader.py` |
| Carrier HTTP | `apps/carrier-service/registry_routes.py` |
| Credentials | `libs/py-core/py_core/credentials_store.py` |
| Compose env | `coding-challenge-2/gondo/docker-compose.yml` |
