# Final Leaderboard: Coding Challenge #1 — All 6 PRs Reviewed

---

## Summary Table

| Rank | Team | PR | Score |
|------|------|----|-------|
| 1 | Thien Nguyen | #2 | 102/120 |
| 2 | NHA | #6 | 95/120 |
| 3 | BSOD | #5 | 94/120 |
| 3 | NoSleep | #7 | 94/120 |
| 5 | Eric | #3 | 93/120 |
| 6 | Hoa Nguyen | #1 | 53/120 |

---

## 1. Thien Nguyen (PR #2) — 102/120

**Ranking Justification:** Thien delivers a polished, library-first solution with excellent separation of concerns (TeamRotator → MemberIterator → Types), comprehensive 31-test suite with explicit fairness verification, and a professional NX monorepo setup with 7 runnable examples. One edge-case crash bug in the iterator prevents a top-tier rating, but overall engineering maturity is high.

**Score Summary:**

| Category | Score | Max |
|----------|-------|-----|
| Core Logic | 44/45 | 45 |
| Code Quality | 22/25 | 25 |
| Tests | 14/15 | 15 |
| Communication | 15/15 | 15 |
| **Subtotal** | **95/100** | **100** |
| Bonus | +7/20 | +20 |
| **Total** | **102/120** | **120** |

---

## 2. NHA (PR #6) — 95/120

**Ranking Justification:** NHA delivers a flawless core logic implementation (45/45) paired with the strongest TDD discipline seen across all submissions — ~86 commits showing textbook Red-Green-Refactor with "fake it till you make it" progression. The 36+ tests (30 unit + 6 acceptance) are comprehensive and well-organized. However, the absence of any formal design pattern, weak documentation, and several code artifacts (dead test files, IntelliJ template) prevent a higher placement.

**Score Summary:**

| Category | Score | Max |
|----------|-------|-----|
| Core Logic | 45/45 | 45 |
| Code Quality | 15/25 | 25 |
| Tests | 14/15 | 15 |
| Communication | 8/15 | 15 |
| **Subtotal** | **82/100** | **100** |
| Bonus | +13/20 | +20 |
| **Total** | **95/120** | **120** |

---

## 3. BSOD (PR #5) — 94/120

**Ranking Justification:** BSOD delivers a technically polished solution with a formal Iterator pattern interface, perfect core logic (45/45), and the strongest documentation/communication of any submission (15/15). The code is well-structured, consistently formatted, and thoroughly commented with JSDoc and complexity annotations. However, the submission **explicitly attributes code generation to Claude 3.5 Sonnet and documentation to Gemini Pro**, which fundamentally questions whether the score reflects the team's engineering capability or the AI's output quality.

**Score Summary:**

| Category | Score | Max |
|----------|-------|-----|
| Core Logic | 45/45 | 45 |
| Code Quality | 21/25 | 25 |
| Tests | 11/15 | 15 |
| Communication | 15/15 | 15 |
| **Subtotal** | **92/100** | **100** |
| Bonus | +2/20 | +20 |
| **Total** | **94/120** | **120** |

---

## 4. NoSleep (PR #7) — 94/120

**Ranking Justification:** NoSleep delivers the most concise and clean implementation reviewed (~50 lines of core code), with correct round-robin logic, explicit no-repetition handling, immutable Member class, and clear TDD evidence (test commit before implementation commit). The code is elegant in its simplicity — no debug artifacts, no dead code, no unnecessary abstractions. The main gaps are minimal test coverage (only 5 tests) and no formal GoF design pattern interface.

**Score Summary:**

| Category | Score | Max |
|----------|-------|-----|
| Core Logic | 43/45 | 45 |
| Code Quality | 19/25 | 25 |
| Tests | 9/15 | 15 |
| Communication | 13/15 | 15 |
| **Subtotal** | **84/100** | **100** |
| Bonus | +10/20 | +20 |
| **Total** | **94/120** | **120** |

---

## 5. Eric (PR #3) — 93/120

**Ranking Justification:** Eric delivers a well-architected solution implementing three formal design patterns (Strategy, Singleton, Factory) with proper interfaces, comprehensive JSDoc documentation, and a professional 398-line README. The core logic is correct and the layered architecture (Controller → Service → Strategy) is exemplary. The main gaps are zero TDD evidence, missing edge case tests, and a few code quality issues (non-null assertion, null-check ordering).

**Score Summary:**

| Category | Score | Max |
|----------|-------|-----|
| Core Logic | 43/45 | 45 |
| Code Quality | 22/25 | 25 |
| Tests | 11/15 | 15 |
| Communication | 14/15 | 15 |
| **Subtotal** | **90/100** | **100** |
| Bonus | +3/20 | +20 |
| **Total** | **93/120** | **120** |

---

## 6. Hoa Nguyen (PR #1) — 53/120

**Ranking Justification:** While the core rotation algorithm (modulo arithmetic) works for the happy path, the submission has a critical bug (`add_member` always increments `active_count` regardless of actual status, causing an infinite loop when inactive members are added), debug artifacts in production code (`print()` on line 23), commented-out code blocks, dead code (`prev_id` is set but never read), no formal design pattern, and only 4 test functions with minimal edge case coverage. The engineering maturity is insufficient for the final round.

**Score Summary:**

| Category | Score | Max |
|----------|-------|-----|
| Core Logic | 30/45 | 45 |
| Code Quality | 10/25 | 25 |
| Tests | 6/15 | 15 |
| Communication | 7/15 | 15 |
| **Subtotal** | **53/100** | **100** |
| Bonus | +0/20 | +20 |
| **Total** | **53/120** | **120** |

---

## Top 3 Recommendation for Finals

1. **Thien Nguyen** (102)
2. **NHA** (95)
3. **BSOD** (94)
