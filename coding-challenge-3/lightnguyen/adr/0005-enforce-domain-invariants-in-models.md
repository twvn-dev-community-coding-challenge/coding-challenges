# ADR-0005: Enforce domain invariants in models

- Status: Accepted
- Date: 2026-10-31

## Context

Domain rules should remain true regardless of which service or API path changes a model. Relying only on callers to validate state can leave a model invalid when it is used elsewhere.

## Decision

Models enforce their own domain invariants and throw an exception when an operation would violate them. Use a specific exception type that describes the failure where practical (for example, `InvalidOperationException` for an invalid state transition). Keep request validation at the API boundary as an additional user-facing check, not as a replacement for domain enforcement.

## Consequences

- Domain behavior remains consistent across callers.
- Callers and tests must account for exceptions from invalid domain operations.
- Exceptions represent invalid operations or state, not routine branching or expected lookup outcomes.
