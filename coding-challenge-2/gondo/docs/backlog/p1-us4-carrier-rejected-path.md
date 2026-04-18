# P1 — User Story 4: Carrier-rejected → retry path

| Field | Value |
|-------|-------|
| **Priority** | P1 |
| **Status** | **Open** |
| **Challenge** | [User Story 4 — SMS Lifecycle](../../../coding-challenge-2.md) |

## Problem

The brief defines:

`Queue → Carrier-rejected → Send-to-provider` (then continuing the flow).

The codebase models **Carrier-rejected** and allows **retry** from **`Send-failed` / `Carrier-rejected`**, but there is **no focused test** that drives a notification through **Queue → Carrier-rejected** and then **retry** to **Send-to-provider**, matching the diagram.

## Value

Demonstrates **controlled transitions** and **traceability** for an edge path called out in the challenge.

## Acceptance criteria

- [ ] Automated test (or documented integration steps) that reaches **Carrier-rejected** from **Queue** via simulated carrier/bus outcome.
- [ ] **Retry** from **Carrier-rejected** transitions toward **Send-to-provider** / **Queue** as designed; assertions on state sequence.
- [ ] Optional: row in [routing-vs-challenge-brief.md](../routing-vs-challenge-brief.md) or lifecycle doc referencing this path.

## References

- `apps/notification-service/models.py` — allowed transitions  
- Challenge: `coding-challenge-2.md` § User Story 4 (transitions)
