# carrier-service

- **HTTP**: `8004` — FastAPI `/health`, **`GET /registry/carriers`** (carrier / MNO YAML registry), and OpenAPI
- **Role**: **Carrier bounded context** — consumes **`sms.dispatch.requested`** from the message bus, performs outbound HTTP to the provider `api_endpoint` from the event (skeleton: `GET`), then publishes **`sms.dispatch.outcome`**. Serves **`infra/carrier-registry/`** (not `infra/provider-registry/`).
- **Rate limiting**: Skeleton concurrency gate — `CARRIER_MAX_CONCURRENT_SENDS` (default `4`); replace with token bucket / per-provider limits as needed.

## Environment

| Variable | Purpose |
|----------|---------|
| `NATS_URL` | Broker URL(s), e.g. `nats://nats:4222` |
| `CARRIER_MAX_CONCURRENT_SENDS` | Max parallel outbound calls |
| `CARRIER_REGISTRY_ROOT` | Root of the carrier YAML tree (default: repo `infra/carrier-registry`; Docker: `/app/infra/carrier-registry`) |
| `CREDENTIALS_BACKEND` / `MOCK_CREDENTIALS_PATH` | Same as provider-service: registry API returns **`carrier_credentials_configured`** without exposing secrets |

## Code layout

- `registry_loader.py` / `registry_routes.py` — load `infra/carrier-registry` and expose `GET /registry/carriers`
- `cqrs/dispatch_handler.py` — event handler + HTTP probe + outcome publish
- `cqrs/rate_gate.py` — `asyncio.Semaphore`
- `lifespan.py` — connect bus + subscribe

## Related docs

Root [README.md](../../README.md) (diagrams), [infra/message-bus/topics.yaml](../../infra/message-bus/topics.yaml).
