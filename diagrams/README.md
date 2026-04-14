# Architecture diagrams (Mermaid)

Diagrams for **[coding-challenge-2/gondo/](../coding-challenge-2/gondo/)** — open any `.md` file on GitHub to render Mermaid.

## Index

| File | What it shows |
|------|----------------|
| [sms-dispatch-sequence.md](sms-dispatch-sequence.md) | **Dispatch E2E:** gRPC **SelectProvider**, **`sms_dispatch_requested_published`**, NATS **requested / received / outcome**, notification states |
| [sms-topics-overview.md](sms-topics-overview.md) | **Three SMS topics:** publisher → subscriber matrix |
| [notification-state-machine.md](notification-state-machine.md) | **Aggregate states**, dispatch **gating** flowchart, **async** Queue → Send-to-carrier |
| [registry-config-loading.md](registry-config-loading.md) | **PROVIDER_REGISTRY_ROOT** vs **CARRIER_REGISTRY_ROOT**, loaders, credentials |
| [carrier-registry-http.md](carrier-registry-http.md) | **`GET /registry/carriers`** (carrier bounded context) |
| [message-bus-gateway-optional.md](message-bus-gateway-optional.md) | **HTTP → NATS** for tooling (**`/v1/publish`**) |

## Single sources of truth in repo

- Message topics & roles: [`coding-challenge-2/gondo/infra/message-bus/topics.yaml`](../coding-challenge-2/gondo/infra/message-bus/topics.yaml)
- Compose ports & env: [`coding-challenge-2/gondo/docker-compose.yml`](../coding-challenge-2/gondo/docker-compose.yml)
