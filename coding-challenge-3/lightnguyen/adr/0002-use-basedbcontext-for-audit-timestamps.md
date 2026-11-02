# ADR-0002: Use `BaseDbContext` for audit timestamps

- Status: Accepted
- Date: 2026-10-31

## Context

Entities inheriting from `BaseModel` need consistent creation and update timestamps without requiring every service to set them manually.

## Decision

Application EF Core contexts inherit from `CodingChallenge.Share.Data.BaseDbContext`. Its `SaveChanges` and `SaveChangesAsync` overrides set `CreatedDateTime` for added entities and `UpdatedDateTime` for modified entities.

## Consequences

- Timestamp handling is centralized and applies to changes saved through either API.
- Services should not set these audit fields themselves during ordinary persistence.
- New contexts must inherit from `BaseDbContext` for this behavior; changes that bypass EF Core saves are outside this automation.
