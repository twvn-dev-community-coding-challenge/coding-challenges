# SMS dispatch sequence (notification → provider → NATS → carrier)

- **YAML / env**: **[registry-config-loading.md](registry-config-loading.md)**
- **States**: **[notification-state-machine.md](notification-state-machine.md)**
- **Topics**: **[sms-topics-overview.md](sms-topics-overview.md)**

**Async integration:** **`sms.dispatch.received`** → **notification-service** (Queue → Send-to-carrier). **`sms.dispatch.outcome`** → **provider-service** (logging only).

NATS subjects: **`gondo.`** + logical topic (`py_core.bus.topics.topic_to_subject`).

```mermaid
sequenceDiagram
    autonumber
    actor Client as Client / UI
    participant NS as notification-service<br/>:8001 HTTP + bus subscriber
    participant DB as PostgreSQL<br/>routing rules + carrier prefix
    participant PS as provider-service<br/>:50051 gRPC / :8002 HTTP
    participant YML as infra/provider-registry<br/>merged YAML
    participant Bus as NATS<br/>gondo.sms.*
    participant CS as carrier-service<br/>:8004
    participant Ext as Provider HTTPS API<br/>api_endpoint from event payload

    Client->>NS: POST /notifications/{id}/dispatch
    NS->>DB: derive carrier from phone prefix
    DB-->>NS: carrier code

    NS->>PS: gRPC SelectProvider(…, policy_context.message_id)
    PS->>DB: resolve routing / select provider
    PS->>YML: yaml_registry — merged provider row
    alt message_id set
        PS->>Bus: publish sms.dispatch.requested
    end
    PS-->>NS: SelectProviderResponse + sms_dispatch_requested_published

    alt sms_dispatch_requested_published = true
        NS->>NS: Send-to-provider → Queue
        NS-->>Client: 200 (state Queue)
    else bus missing / publish skipped
        NS-->>Client: 503 DISPATCH_REQUEST_NOT_PUBLISHED (state stays New)
    end

    Note over Bus,NS: Below: async path after successful 200 + Queue

    Bus-->>CS: sms.dispatch.requested
    CS->>CS: acquire concurrency slot
    CS->>Bus: publish sms.dispatch.received
    Bus-->>NS: sms.dispatch.received
    NS->>NS: Queue → Send-to-carrier

    CS->>Ext: GET api_endpoint (from event; originally from provider registry at publish)
    Ext-->>CS: HTTP response
    CS->>Bus: publish sms.dispatch.outcome
    Bus-->>PS: log outcome
```

## Notes

| Topic | Detail |
|-------|--------|
| **Queue vs Send-to-carrier** | HTTP **200** returns **Queue**; **Send-to-carrier** appears after **`sms.dispatch.received`** (async). |
| **503 on dispatch** | **`sms_dispatch_requested_published`** false → no move to Queue; aggregate stays **New**. |
| **Callbacks** | **`POST /provider-callbacks`** — external provider webhooks (not shown). |

## Code references

| Step | Location |
|------|----------|
| YAML registries | **[registry-config-loading.md](registry-config-loading.md)** |
| `sms_dispatch_requested_published` | `libs/grpc-contracts/protos/provider.proto`; `apps/provider-service/grpc_server.py`; `cqrs/publish_dispatch.py` |
| Dispatch + 503 guard | `apps/notification-service/main.py` |
| gRPC client | `apps/notification-service/grpc_client.py` |
| Topics | `libs/py-core/py_core/bus/topics.py`, `infra/message-bus/topics.yaml` |
| Carrier handler | `apps/carrier-service/cqrs/dispatch_handler.py` |
| Received → state | `apps/notification-service/cqrs/dispatch_received_subscriber.py`, `cqrs/carrier_dispatch_received.py` |
| Outcome logs | `apps/provider-service/cqrs/outcome_subscriber.py` |
