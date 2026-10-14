# message-bus-gateway

- **HTTP**: `8010`
- **Role**: Optional **sidecar** for operators and scripts: **`POST /v1/publish`** forwards JSON to NATS using the same subject mapping as `py_core.bus.topics.topic_to_subject` (`gondo.<logical-topic>`).
- **Not required** for provider or carrier in production — those use the in-process NATS client via `libs/py-core`.

## Environment

| Variable | Purpose |
|----------|---------|
| `NATS_URL` | Broker URL(s) |

## Endpoints

- `GET /health` — liveness
- `POST /v1/publish` — body `{ "topic": "sms.dispatch.requested", "payload": { ... } }`

## Related docs

[infra/message-bus/topics.yaml](../../infra/message-bus/topics.yaml), root [README.md](../../README.md).
