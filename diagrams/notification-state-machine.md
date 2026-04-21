# Notification lifecycle (aggregate)

String states and **`VALID_TRANSITIONS`** live in **`apps/notification-service/models.py`**.

- **Queue → Send-to-carrier** is **async** via NATS **`sms.dispatch.received`** (`cqrs/carrier_dispatch_received.py`).
- **New → Queue** on dispatch requires **gRPC `sms_dispatch_requested_published`**; otherwise **HTTP 503** and state stays **New** (see dispatch sequence).

## Dispatch gating (before Queue)

```mermaid
flowchart TD
  N[New]
  N -->|POST …/dispatch| G{SelectProvider OK?}
  G -->|no| E[502 / error]
  G -->|yes| P{sms_dispatch_requested_published?}
  P -->|no| N
  P -->|yes| ST[Send-to-provider → Queue]
```

## Aggregate states (valid transitions)

```mermaid
stateDiagram-v2
    [*] --> New

    New --> Send_to_provider: dispatch success + published
    Send_to_provider --> Queue: same HTTP handler

    Queue --> Send_to_carrier: NATS sms.dispatch.received

    Send_to_carrier --> Send_success: provider-callbacks
    Send_to_carrier --> Send_failed: provider-callbacks
    Send_to_carrier --> Carrier_rejected: provider-callbacks OR sim MNO reject (090909094)

    Send_failed --> Send_to_provider: retry
    Carrier_rejected --> Send_to_provider: retry

    Send_success --> [*]
```

Challenge brief also lists **`Queue → Carrier-rejected`** without **Send-to-carrier** (shorthand). This repo treats **Carrier-rejected** per definition as MNO rejection after provider acceptance: **`Queue → Send-to-carrier`** on **`sms.dispatch.received`**, then **`Send-to-carrier → Carrier-rejected`** when the simulated MNO refuses the MT (see **`carrier_auto_reject`**), or **`POST /provider-callbacks`** with **`new_state`** **`Carrier-rejected`** from **Send-to-carrier**.

Code uses hyphenated names: **`Send-to-provider`**, **`Queue`**, **`Send-to-carrier`**, **`Send-success`**, **`Send-failed`**, **`Carrier-rejected`**.

## Code references

| Area | Path |
|------|------|
| Transitions table | `apps/notification-service/models.py` |
| Dispatch / retry + publish guard | `apps/notification-service/main.py` |
| Async Queue → Send-to-carrier | `apps/notification-service/cqrs/carrier_dispatch_received.py`, `cqrs/dispatch_received_subscriber.py` |
| Provider callbacks | `POST /provider-callbacks` in `main.py` |
