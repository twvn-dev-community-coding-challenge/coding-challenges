# Message bus gateway (optional)

Production services use **`py_core.bus.MessageBus`** (NATS) in-process. The gateway lets **scripts or tools** publish without **`nats-py`**: **`POST /v1/publish`** with a **logical topic** and **JSON payload** — subject becomes **`gondo.<topic>`**.

```mermaid
sequenceDiagram
    autonumber
    participant Client as HTTP client
    participant GW as message-bus-gateway :8010
    participant Bus as NATS

    Note over GW: Same subject rules as py_core.bus.topics.topic_to_subject
    Client->>GW: POST /v1/publish
    Note right of Client: body: topic, payload object
    GW->>Bus: publish(subject, JSON bytes)
    Bus-->>GW: ok
    GW-->>Client: status + subject
```

## Compared to in-process bus

| Path | Use case |
|------|----------|
| **Gateway** | curl / Postman / CI publishing test events |
| **py_core.bus** in apps | provider, carrier, notification subscribers |

## Code references

- `apps/message-bus-gateway/main.py` — `/v1/publish`, `/health`
- `libs/py-core/py_core/bus/topics.py` — `topic_to_subject`
