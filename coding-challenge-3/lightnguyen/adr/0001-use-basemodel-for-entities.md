# ADR-0001: Use `BaseModel` for persisted models

- Status: Accepted
- Date: 2026-10-31

## Context

New persisted models need a consistent identity and common lifecycle fields. Repeating these fields on every model adds boilerplate and makes it easier for entities to diverge.

## Decision

Persisted models inherit from `CodingChallenge.Share.Models.BaseModel`. It provides a generated `Guid` identifier, soft-delete flag, and creation and update timestamps. Model-specific properties remain on the derived model.

## Consequences

- New models get the common fields by default and can use the shared persistence conventions.
- The base class is intended for persisted models; request and response DTOs should not inherit from it.
- If a model requires different identity or lifecycle semantics, that exception should be explicit.
