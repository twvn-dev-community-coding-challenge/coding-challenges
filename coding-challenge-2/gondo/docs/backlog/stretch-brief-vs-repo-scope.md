# Stretch — Challenge brief vs repo scope (reviewer note)

| Field | Value |
|-------|-------|
| **Priority** | Stretch |
| **Status** | **Partial** — `SUBMISSION.md` § Trade-offs & limitations already contrasts sim stack vs production; add one explicit sentence tying to **official brief** (“CLI/in-memory”) if reviewers want literal mapping |
| **Challenge** | [Technical expectations & out-of-scope](../../../coding-challenge-2.md) |

## Problem

The official brief suggests **CLI or simple triggers**, **in-memory storage**, and **no database persistence** for the exercise. This repository intentionally goes further: **REST APIs**, **PostgreSQL**, **NATS**, **frontend**, **OpenAPI**, **otp-service**, etc.

Judges may ask how the submission maps to the brief.

## Value

One short, honest narrative avoids confusion without refactoring the solution.

## Acceptance criteria

- [x] Narrative exists: **`SUBMISSION.md`** § Trade-offs and limitations + design decisions cover REST/Postgres/NATS/UI vs simulation; optionally rename subsection to **“Challenge brief vs this repository”** for judges who skim only headings.
- [x] Link [system-vs-program-requirements.md](../system-vs-program-requirements.md) judge quickstart + backlog [challenge-vs-repo-gap-scan.md](challenge-vs-repo-gap-scan.md).

## References

- `coding-challenge-2.md` § Expected Deliverables / Out of Scope  
- Main `README.md` — Minimal deployment
