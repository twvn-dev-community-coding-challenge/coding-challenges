# ADR-0007: Handle duplicate requests and events idempotently

- Status: Accepted
- Date: 2026-10-31

## Context

HTTP requests and RabbitMQ deliveries can be retried. Processing the same ride creation or driver action more than once must not create duplicate rides, offers, or invalid lifecycle transitions.

## Decision

- Ride creation requires an `Idempotency-Key` header. The ride service stores the key, a hash of the request, and the original response in the same database save as the ride. Repeating the same request with that key returns the stored result; reusing the key with a different request is rejected. The key's primary-key constraint arbitrates concurrent requests, and a losing request reloads and returns the winner's record.
- A retry of ride creation republishes `RideCreated` unless the ride has been cancelled. This allows a client retry to recover if saving succeeded but event publication did not. Consumers and ride state handlers therefore tolerate redelivery.
- Driver acceptance is idempotent when its offer is already accepted: the existing offer is returned without changing driver state again.
- Ride event handlers apply transitions only from their expected source state. Repeated or stale action-completion and driver-matched events are no-ops; terminal rides are not reopened.

## Consequences

- Delivery remains at-least-once; event IDs are not separately recorded. Idempotency comes from persisted request keys and guarded domain state transitions.
- A reused key with a different request is a client error and must be retried with a new key.
- Tests cover repeated ride creation, driver acceptance, and ride state-transition events.
