# Stretch — Challenge brief vs repo scope (reviewer note)

| Field | Value |
|-------|-------|
| **Priority** | Stretch |
| **Status** | **Open** (documentation) |
| **Challenge** | [Technical expectations & out-of-scope](../../../coding-challenge-2.md) |

## Problem

The official brief suggests **CLI or simple triggers**, **in-memory storage**, and **no database persistence** for the exercise. This repository intentionally goes further: **REST APIs**, **PostgreSQL**, **NATS**, **frontend**, **OpenAPI**, **otp-service**, etc.

Judges may ask how the submission maps to the brief.

## Value

One short, honest narrative avoids confusion without refactoring the solution.

## Acceptance criteria

- [ ] Add a subsection to **`SUBMISSION.md`** or **`README.md`** (“Challenge scope vs our build”) stating: REST/UI/DB are **deliberate** for platform-SMS storytelling; core **domain** behaviors still match routing, lifecycle, cost, callbacks.
- [ ] Link [system-vs-program-requirements.md](../system-vs-program-requirements.md) **Minimal deployment** / anti–over-engineering section.

## References

- `coding-challenge-2.md` § Expected Deliverables / Out of Scope  
- Main `README.md` — Minimal deployment
