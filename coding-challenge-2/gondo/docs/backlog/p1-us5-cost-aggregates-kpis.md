# P1 — User Story 5: cost & volume KPIs (aggregates)

| Field | Value |
|-------|-------|
| **Priority** | P1 |
| **Status** | **Open** |
| **Challenge** | [User Story 5 — Cost Tracking & Observability](../../../coding-challenge-2.md) |

## Problem

The brief requires **aggregate** visibility, not only per-message costs:

- Total cost **per provider**
- Total cost **per country**
- SMS **volume** per provider
- **Success / failure** rates

Today the build exposes **estimate/actual on each notification** and charging records per message; **rolled-up** reporting is not a first-class API or export.

## Value

Closes the explicit **User Story 5** acceptance criteria and supports evaluation **§ Observability & Traceability**.

## Acceptance criteria

- [ ] Define surface area: read model API, CLI, SQL views over existing stores, or documented “how to query” for reviewers.
- [ ] Totals (or time-bounded aggregates) **by provider** and **by country** align with stored estimates/actuals.
- [ ] **Volume** and **success/failure** rates are derivable from notification states (or charging callbacks) with a repeatable recipe or automation.

## References

- Challenge: `coding-challenge-2.md` § User Story 5  
- [p1-cost-story.md](p1-cost-story.md) (per-message cost wiring)
