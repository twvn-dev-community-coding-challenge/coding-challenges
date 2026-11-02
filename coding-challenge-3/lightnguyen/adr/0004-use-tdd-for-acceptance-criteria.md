# ADR-0004: Use TDD to implement acceptance criteria

- Status: Accepted
- Date: 2026-10-31

## Context

Acceptance criteria describe observable behavior. Implementing them without corresponding tests risks missing edge cases and makes regressions harder to detect.

## Decision

Use a test-driven development (TDD) approach for acceptance criteria: write a focused test for the expected behavior, run it to confirm it fails for the intended reason, implement the behavior, and run the test again. Add tests for relevant validation and failure cases as well as successful behavior.

## Consequences

- Tests provide executable documentation of acceptance behavior.
- Changes can be verified and refactored with a focused regression suite.
- Tests should assert observable outcomes rather than implementation details wherever practical.
