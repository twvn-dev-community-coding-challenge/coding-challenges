# ADR-0008: Prevent concurrent driver assignments

- Status: Accepted
- Date: 2026-10-31

## Context

Ride-created events can be processed concurrently. Two matching workers can both observe the same driver as available before either creates an offer, allowing one driver to receive offers for multiple rides.

## Decision

SQLite enforces a unique partial index on `Offers.DriverId` for pending and accepted offers. The index is the final arbiter when concurrent workers attempt assignment. The driver service treats that specific unique-constraint conflict as an unsuccessful offer attempt, and the ride-created consumer publishes a driver-matched event without a driver so the ride remains in matching.

Completed offers leave the active set. Rejected, cancelled, and timed-out offers are already outside the index.

## Consequences

- A driver can have at most one pending offer or accepted ride, including when separate processes race.
- The losing ride remains matchable instead of being assigned to a driver who already holds another offer.
- Startup creates the index if it is missing, including on databases created before this change. Startup fails explicitly if existing active offers already violate the constraint, so those records must be resolved before deploying.
