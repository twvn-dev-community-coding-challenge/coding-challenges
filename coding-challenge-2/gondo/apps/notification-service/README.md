# notification-service

- **HTTP**: `8001` — SMS notification lifecycle API (create, dispatch, retry, provider callbacks).
- **gRPC clients**: `provider-service` (`PROVIDER_GRPC_TARGET`); `charging-service` (`CHARGING_GRPC_TARGET`) for **`EstimateCost`** on dispatch/retry and **`RecordActualCost`** on **Send-success** / **Send-failed** callbacks.
- **Message bus**: Subscribes to **`sms.dispatch.received`** (`NATS_URL`) and applies **Queue → Send-to-carrier** asynchronously when **carrier-service** publishes the ack (no sync HTTP from provider).
- **Storage**: In-memory notification store (simulation); PostgreSQL used for carrier prefix resolution.

For OpenAPI generation, NX targets, and end-to-end flows, see the repository root [README.md](../../README.md).
