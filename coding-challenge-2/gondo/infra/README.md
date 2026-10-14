# Infrastructure compose fragments

All paths in `infra/docker-compose.*.yml` are **relative to each file’s directory** (`gondo/infra/`). Run commands from **`gondo/`** (repository app root) so `context: ..` in standalone files resolves to `gondo/`.

## Main stack

| File | Purpose |
|------|---------|
| [`../docker-compose.yml`](../docker-compose.yml) | Full stack: Postgres, NATS, **WireMock** (carrier HTTP probe), charging, provider, carrier, notification, frontend, gateway |

```bash
cd gondo
docker compose up -d --build
```

## Overlays (merge with main compose)

Use **multiple `-f` flags**; later files override / extend earlier ones.

| File | What it adds |
|------|----------------|
| [`docker-compose.provider-registry.yml`](docker-compose.provider-registry.yml) | Bind-mount `infra/provider-registry` into **provider-service** |
| [`docker-compose.carrier-registry.yml`](docker-compose.carrier-registry.yml) | Bind-mount `infra/carrier-registry` into **carrier-service** |
| [`docker-compose.mock-credentials.yml`](docker-compose.mock-credentials.yml) | Bind-mount `infra/credentials/mock` into **provider-service** and **carrier-service** |
| [`docker-compose.dev-registries.yml`](docker-compose.dev-registries.yml) | All of the above in one overlay |

Example (edit YAML + secrets without rebuilding images):

```bash
cd gondo
docker compose -f docker-compose.yml -f infra/docker-compose.dev-registries.yml up -d
```

## Minimal stacks (infra-only pieces)

| File | Services | Typical use |
|------|------------|-------------|
| [`docker-compose.postgres.yml`](docker-compose.postgres.yml) | Postgres only | Run Python apps on host (`yarn nx run …:serve`), DB in Docker |
| [`docker-compose.nats.yml`](docker-compose.nats.yml) | NATS only | Broker for local bus experiments |
| [`docker-compose.messaging.yml`](docker-compose.messaging.yml) | NATS + **message-bus-gateway** | HTTP → NATS without the rest of the stack |

```bash
cd gondo
docker compose -f infra/docker-compose.postgres.yml up -d
docker compose -f infra/docker-compose.messaging.yml up -d --build
```

## Optional cloud mock

Floci (local AWS-compatible API) is defined in the **main** compose file with profile `mock-cloud`. See root [`README.md`](../README.md).
