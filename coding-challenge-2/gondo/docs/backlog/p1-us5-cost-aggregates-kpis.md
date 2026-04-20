# P1 — User Story 5: cost & volume KPIs (aggregates)

| Field | Value |
|-------|-------|
| **Priority** | P1 |
| **Status** | **Done** |
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

- [x] **Read model API:** `GET /notifications/kpis` aggregates **in-memory** notifications (`kpis.py` + `store.list_notifications()`).
- [x] Totals **by provider** and **by country:** `estimated_cost` / `last_actual_cost` sums; volume = row counts.
- [x] **Volume** and terminal **success/failure** rates from `state` (Send-success vs Send-failed + Carrier-rejected).
- [x] **UI:** `/kpis` — **SMS KPIs** page (`SmsKpisPage`).

## References

- Challenge: `coding-challenge-2.md` § User Story 5  
- [p1-cost-story.md](p1-cost-story.md) (per-message cost wiring)
- `apps/notification-service/kpis.py`, `GET /notifications/kpis` in `main.py`
- `apps/frontend/src/pages/sms-kpis/sms-kpis-page.tsx`, route `/kpis`
