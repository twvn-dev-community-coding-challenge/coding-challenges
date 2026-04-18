# P1 — Routing seeds vs User Story 3

| Field | Value |
|-------|-------|
| **Priority** | P1 |
| **Status** | **Done** — [`routing-vs-challenge-brief.md`](../routing-vs-challenge-brief.md) marks **Story 3** PH + VN rows **Match · Yes** against **`0002_seed_data.py`** |
| **Area** | Core correctness — PH + VN routing |

## Problem

**User Story 3** specifies concrete routing scenarios (Philippines rollout, updated Vietnam rules). Database seeds and migrations must **match** the challenge tables line by line.

## Value

Demonstrates operational understanding of **country + carrier** rules and **as_of** rollout.

## Acceptance criteria

- [x] Story 3 VN + PH carrier rows reconciled vs seeds (see **`routing-vs-challenge-brief.md`** tables).
- [x] Doc includes refresh instruction when **`0002_seed_data.py`** changes.
- [x] README / **routing-vs-challenge-brief** narrative covers **time-effective** rules (`as_of`, version **2** from **2026-04-01** vs earlier window — see routing doc § Vietnam).

## References

- [routing-vs-challenge-brief.md](../routing-vs-challenge-brief.md)
- [system-vs-program-requirements.md § User Story 3](../system-vs-program-requirements.md#user-story-3--ph-market--vn-rule-updates)
