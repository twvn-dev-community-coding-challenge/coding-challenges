# Specs-Kit: SMS Message Status — Invalid Transitions and Recovery

This document summarizes **when** an SMS message can hit an **unexpected invalid transition** (relative to `MessageStatus.IsValidTransition` in `internal/sms/models.go`) and **how** to recover operationally and in code.

Related: [SPECS_KIT_TESTCASES_MATRIX.md](./SPECS_KIT_TESTCASES_MATRIX.md) (Section 3, FSM-INV-*), [SOLUTION_DIAGRAMS.md](../SOLUTION_DIAGRAMS.md) (lifecycle).

---

## 1. Definition

An **invalid transition** occurs when `DefaultSMSService.updateStatus` is asked to move from current `msg.Status` to `newStatus` where `IsValidTransition(current, new)` is **false**. The service returns an error **before** persisting `newStatus`, so the **database row keeps the previous valid status** (no partial state write for that failed step).

---

## 2. Allowed transitions (reference)


| Current status     | Allowed next statuses                                |
| ------------------ | ---------------------------------------------------- |
| `New`              | `Send-to-provider`, `Send-failed`                    |
| `Send-to-provider` | `Queue`, `Send-failed`                               |
| `Queue`            | `Send-to-carrier`, `Carrier-rejected`, `Send-failed` |
| `Send-to-carrier`  | `Send-success`, `Send-failed`                        |
| `Carrier-rejected` | `Send-to-provider`                                   |
| `Send-failed`      | `Send-to-provider`                                   |
| `Send-success`     | *(terminal — no further transitions)*                |


---

## 3. Unexpected invalid transition scenarios (summary)


| #   | Scenario                                     | Typical cause                                                       | Example invalid edge                                                   |
| --- | -------------------------------------------- | ------------------------------------------------------------------- | ---------------------------------------------------------------------- |
| A   | **Out-of-order webhook**                     | Provider or adapter sends a “late” or “early” status vs local state | `Queue` → `Send-success` (skips `Send-to-carrier`)                     |
| B   | **Backward jump**                            | Duplicate or replayed callback                                      | `Send-to-carrier` → `Queue`                                            |
| C   | **Post-terminal update**                     | Second delivery receipt after success                               | `Send-success` → `Send-failed` or any non-identity change              |
| D   | **Callback while still in early sync phase** | Webhook arrives before `SendSMS` finished persisting expected state | e.g. targeting `Send-success` while still `New` / `Send-to-provider`   |
| E   | **Corrupt or manual DB state**               | Migration, manual SQL, bug elsewhere                                | Any transition from a state that does not match real provider timeline |
| F   | **Wrong status literal**                     | Client posts wrong `status` JSON value                              | Maps to disallowed edge from current row                               |


**Note:** The synchronous `SendSMS` path is written to follow valid edges. Invalid transitions in production are **most often** from `**HandleCallback`** accepting arbitrary `status` without normalizing to the current FSM state.

---

## 4. Recovery methods (summary)


| Method                               | When to use                                     | Action                                                                                                                                          |
| ------------------------------------ | ----------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| **R1 — No-op on error**              | Invalid transition returned from `updateStatus` | Treat as **safe**: row unchanged; log correlation id + provider `message_id` + attempted transition; **do not** retry same payload blindly.     |
| **R2 — Webhook normalization**       | Steady-state integration with CPaaS             | Map each provider event to **either** the single next allowed state **or** ignore if already past that state (idempotency).                     |
| **R3 — Ordered replay**              | Out-of-order delivery                           | Buffer or reorder by provider timestamp **or** apply only if `from → to` is valid; drop invalid.                                                |
| **R4 — Idempotent terminal**         | Success already recorded                        | If current status is `Send-success`, accept duplicate success notifications without calling `updateStatus`.                                     |
| **R5 — Manual reconciliation**       | DB/provider disagree                            | Use `status_logs` + provider console; **manually** set `status` to a state consistent with truth, then resume callbacks only along valid edges. |
| **R6 — Admin repair API** (optional) | Operations need correction                      | Endpoint that validates `from → to` before `UPDATE`, or emits alert instead of silent fix.                                                      |


---

## 5. Test case mapping (specs-kit)


| Spec ID                   | Focus                                                                                                               |
| ------------------------- | ------------------------------------------------------------------------------------------------------------------- |
| SMS-LC-005                | Skip step: `Queue` → direct terminal-like status → expect error, no invalid persist                                 |
| SMS-LC-006                | Terminal: further callbacks rejected                                                                                |
| FSM-INV-001 … FSM-INV-004 | See [SPECS_KIT_TESTCASES_MATRIX.md](./SPECS_KIT_TESTCASES_MATRIX.md#3-message-status-transitions-isvalidtransition) |


---

## 6. Automatic recovery inside service (today)

When status moves to `**Carrier-rejected`** or `**Send-failed**` via valid transitions, `updateStatus` may chain a further transition to `**Send-to-provider**` (fallback). That path only runs **after** a **valid** transition into those failure states; it does **not** fix invalid webhook jumps.

---

## Revision history


| Version | Date       | Notes                         |
| ------- | ---------- | ----------------------------- |
| 1.0     | 2026-04-10 | Initial summary for specs-kit |


