# ADR-0003: Use SQLite for quick application setup

- Status: Accepted
- Date: 2026-10-31

## Context

The challenge applications need a relational database that is quick to configure and run locally without requiring a separately installed database server.

## Decision

Use SQLite as the default application database during challenge development. Configure it through EF Core with a local database file.

## Consequences

- Developers can run the application with minimal setup.
- SQLite is a relational provider, so common EF Core mapping and persistence behavior can be exercised locally.
- Provider-specific behavior and deployment requirements must be considered if the application later moves to another database; verify migrations and integration behavior against the target provider.
