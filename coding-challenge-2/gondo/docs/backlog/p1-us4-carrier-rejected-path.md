# P1 — User Story 4: Carrier-rejected → retry path

| Field | Value |
|-------|-------|
| **Priority** | P1 |
| **Status** | **Done** — `cqrs/carrier_auto_reject.py` + `apply_carrier_dispatch_received`; tests `test_carrier_rejected_e2e.py`, `test_carrier_auto_reject.py` |
| **Challenge** | [User Story 4 — SMS Lifecycle](../../../coding-challenge-2.md) |

## Problem

The brief defines:

`Queue → Carrier-rejected → Send-to-provider` (then continuing the flow).

The codebase models **Carrier-rejected** and allows **retry** from **`Send-failed` / `Carrier-rejected`**, but there is **no focused test** that drives a notification through **Queue → Carrier-rejected** and then **retry** to **Send-to-provider**, matching the diagram.

## Value

Demonstrates **controlled transitions** and **traceability** for an edge path called out in the challenge.

## Acceptance criteria

- [x] Automated test: **`sms.dispatch.received`** → **Queue → Send-to-carrier**, then **Send-to-carrier → Carrier-rejected** for **VN** **`090909094`** / **`+8490909094`** (simulated MNO rejection per challenge definition).
- [x] **Retry** from **Carrier-rejected** → **Send-to-provider** → **Queue**; second bus event can reject again (same test MSISDN).
- [ ] Optional: lifecycle diagram cross-link (see `carrier_dispatch_received.py`).

## References

- `apps/notification-service/models.py` — allowed transitions  
- `apps/notification-service/cqrs/carrier_auto_reject.py` — test MSISDN **`8490909094`** (`090909094`)
- `apps/notification-service/cqrs/carrier_dispatch_received.py` — **`sms.dispatch.received`** consumer
- Challenge: `coding-challenge-2.md` § User Story 4 (transitions)
