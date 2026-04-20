# Coding Challenge Program — Understanding (TWVN Coding Challenge 2026)

This note captures what the official **program overview** asks for — **process, scoring, submission** — independent of the **challenge-specific** functional brief ([`coding-challenge-2.md`](../coding-challenge-2.md) — sibling of the `gondo/` folder).

**Related:** [System vs program mapping & improvements](system-vs-program-requirements.md) (how `gondo` lines up and what to improve).

---

## Purpose

- Internal learning event: experiment, pair across teams; **partial solutions are explicitly welcome**.
- Deliverables are **code + communication** (README, PR, optional Demo Day), not only features.

## Participation model

| Aspect | Detail |
|--------|--------|
| Who | Solo or teams of **1–3** |
| Time | Roughly **2–3 hours/week** over **2–3 weeks** per challenge; FAQ suggests **~4–6 hours total** as a sanity check |
| Where | Community channel (`#Vietnam Dev Community`), organizers/mentors |

## Submission mechanics (what “done” means administratively)

| Step | Action |
|------|--------|
| 1 | **Fork** the official challenge repository; implement in **your fork**. |
| 2 | Branch: `coding-challenge-{number}/<team-name>` (e.g. `coding-challenge-2/gondo`). |
| 3 | Put **all** artifacts under `coding-challenge-{number}/<team-name>/` — code, tests, config, **README** (template sections). |
| 4 | Open a **Pull Request** to the **official** repo from your fork (required when possible). **Do not self-merge.** |
| 5 | Fallback: share **fork URL + branch** with organizers if PR is impossible. |

PR body should cover: team, overview, design decisions, stack, **how to run / test**, trade-offs, **with more time**, **AI tools** (if any).

## Rules of the road

| ✅ Allowed | ❌ Not allowed |
|------------|----------------|
| Any reasonable stack unless the challenge specifies | Client or proprietary data |
| Open-source libraries | **Over-engineering beyond scope** |
| AI assistants **if documented** (may count toward bonus) | |

## Scoring framework (generic — every challenge)

Per challenge: **100 points + bonuses** (challenge-specific rubric fills in the **70-point** “core” bucket).

| Block | Points | What judges look for |
|-------|--------|----------------------|
| **A. Core execution & design** | **70** | Correctness + architecture — **details come from each challenge doc**, not from the program overview alone. |
| **B. Tests & reliability** | **15** | Coverage of core logic (**10**) + edge cases / errors (**5**). |
| **C. Documentation & communication** | **15** | README / setup / trade-offs (**5**) + explaining decisions (**10**, often reinforced on Demo Day). |

**Bonuses** (examples from program text): TDD (+5), junior-led demo (+5), peer reviews (+5), creative (+10).

**Season:** Four challenges; **season score** = sum of **top three** scores (**best 3 of 4**). Per-challenge podium prizes are separate from the season prize narrative.

## Typical Demo Day arc (~10–15 min)

Problem recap → live demo or code walkthrough → design & trade-offs → lessons learned. A failed demo with a clear explanation often scores better than silence.

## FAQ (engineering-relevant)

- Juniors encouraged; learning and honest “we tried X” narratives count.
- Partial beats nothing.
- Reuse **ideas**, not entire unrelated codebases.
- Document AI assistance if you use it.

## What this document does *not* define

| Topic | Where it lives |
|-------|----------------|
| OTP vs SMS ownership, routing tables, Philippines rollout, lifecycle states | **[`coding-challenge-2.md`](../coding-challenge-2.md)** (challenge brief) |
| Exact sub-criteria inside the **70** points | Challenge facilitator / challenge-specific scoring sheet |

---

*Synthesized from the TWVN Coding Challenge 2026 program overview text.*
