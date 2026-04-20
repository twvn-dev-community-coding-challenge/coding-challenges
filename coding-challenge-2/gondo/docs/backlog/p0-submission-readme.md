# P0 — Submission template & README hygiene

| Field | Value |
|-------|-------|
| **Priority** | P0 |
| **Status** | **Done (minor gaps)** — `SUBMISSION.md` and main `README.md` cover template sections; **`AI tools used`** may still need content if assistants were used |
| **Area** | Documentation / program scoring |

## Problem

Organizers expect a clear **submission narrative**: approach, decisions, how to run, how to test, trade-offs, AI disclosure. Scattered or missing sections hurt the **documentation & communication** scoring block.

## Value

Reviewers can validate the solution quickly without reverse-engineering the repo.

## Acceptance criteria

- [x] `SUBMISSION.md` fills every organizer template section (or explicitly defers with reason).
- [x] Main `README.md` links to submission doc and diagrams; submission checklist mapped via **For reviewers** + `SUBMISSION.md` link.
- [x] PR description can be assembled from README + `docs/system-vs-program-requirements.md` snippets (incl. PR snippet in §1).
- [ ] Optional: **`AI tools used`** in `SUBMISSION.md` completed when applicable.

## References

- [system-vs-program-requirements.md § 1](../system-vs-program-requirements.md#1-program-scenario-fork-branch-folder-pr)
- Repo root `SUBMISSION.md`, `README.md`
