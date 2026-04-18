# P0 — Judge quickstart & limitations

| Field | Value |
|-------|-------|
| **Priority** | P0 |
| **Status** | **Done** — README **For reviewers**, **Scope and limitations**, **Tests**, Docker notes |
| **Area** | Documentation |

## Problem

Judges need a **copy-paste quickstart** and an honest **limitations** paragraph (simulated SMS, in-memory notifications, etc.) without reading the entire README.

## Value

Faster review, fewer environment mistakes, clearer boundaries between demo and production-grade behavior.

## Acceptance criteria

- [x] Quickstart commands documented (**For reviewers** + install/migrate/start); validated by team workflow.
- [x] Limitations stated (**Scope and limitations**): simulated SMS, in-memory notifications, OTP demo flag note, NATS/docker context where relevant.
- [x] Test command + Postgres / `DATABASE_URL` note (`README.md`, **`SUBMISSION.md` — How to test**, `system-vs-program-requirements` Judge quickstart).

## References

- [system-vs-program-requirements.md § Judge quickstart](../system-vs-program-requirements.md#judge-quickstart-copy-paste)
- Main `README.md` — Scope and limitations
