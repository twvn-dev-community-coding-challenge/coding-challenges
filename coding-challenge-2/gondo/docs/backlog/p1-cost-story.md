# P1 — Cost estimate & actual wired end-to-end

| Field | Value |
|-------|-------|
| **Priority** | P1 |
| **Status** | **Done** — README **Cost tracking** describes `EstimateCost` / `RecordActualCost`; notification tests include charging failures (`CHARGING_RATE_NOT_FOUND`, etc.) |
| **Area** | Challenge brief — lifecycle & charging |

## Problem

The brief ties **estimated cost** to moving into Send-to-provider and **actual cost** to Send-success (and related lifecycle points). Any gap between story and observable API behavior should be closed or explicitly documented.

## Value

Demonstrates integration with **charging-service** and satisfies core narrative for cost-aware SMS orchestration.

## Acceptance criteria

- [x] Dispatch/retry estimate path aligned with README (`EstimateCost`, persisted refs on notification resource).
- [x] Callback paths call **`RecordActualCost`** where design specifies (success/failed with provider selected).
- [x] Notification JSON exposes charging-related fields, a read-only **`cost_story`** (brief: estimate at Send-to-provider, actual at callback), and error coverage for **`CHARGING_RATE_NOT_FOUND`** / gRPC paths (`test_cost_story.py`, `test_dispatch.py`, `test_callbacks.py`).
- [x] Simulation boundaries called out in **Scope and limitations** + **Cost tracking** (no live CPaaS; charging is real gRPC to **charging-service**).

## References

- [system-vs-program-requirements.md § 4](../system-vs-program-requirements.md#4-cost-story-challenge-estimate-vs-actual)
- Main `README.md` — Cost tracking
