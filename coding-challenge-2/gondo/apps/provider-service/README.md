# provider-service

- **HTTP**: `8002` (FastAPI health + OpenAPI)
- **gRPC**: `50051` — `ProviderService` (`ResolveRouting`, `SelectProvider`, `GetProviderRegistry`)
- **Data**: PostgreSQL routing rules; **provider YAML registry** at `PROVIDER_REGISTRY_ROOT` (default `../../infra/provider-registry` in dev, `/app/infra/provider-registry` in Docker). Registry is **partitioned by country**: `countries/<ISO>/` indexes + provider fragments; the loader **merges** the same `provider_id` across countries (e.g. Twilio in VN + SG). MNO-level registry lives in **`infra/carrier-registry/`** and is exposed by **carrier-service** only (`GET /registry/carriers`).
- **Credentials:** provider YAML `credentials_ref` maps to entries under `infra/credentials/mock/secrets.json` when `CREDENTIALS_BACKEND=mock`, or to **AWS Secrets Manager** when `CREDENTIALS_BACKEND=aws_secrets_manager` and `AWS_ENDPOINT_URL` points at [Floci](https://floci.io/) (`docker compose --profile mock-cloud`). gRPC only returns **whether** a secret is configured, never raw values.

## CQRS + message bus

On startup the process connects to **NATS** (`NATS_URL`) and:

1. Subscribes to `sms.dispatch.outcome` (logical topic; NATS subject `gondo.sms.dispatch.outcome`).
2. After a successful `SelectProvider`, if `policy_context.message_id` is set, publishes `sms.dispatch.requested` with the registry `api_endpoint` and ids. In Docker Compose, **`CARRIER_HTTP_PROBE_URL`** overrides that URL so **carrier-service** can `GET` a **WireMock** stub (`infra/wiremock/mappings/`) instead of calling real provider APIs.

Code layout: `cqrs/`, `lifespan.py` (gRPC + bus), `bus_state.py`, `yaml_registry.py`.

## Related docs

See the root [README.md](../../README.md) (Message bus section) and [infra/message-bus/topics.yaml](../../infra/message-bus/topics.yaml).
