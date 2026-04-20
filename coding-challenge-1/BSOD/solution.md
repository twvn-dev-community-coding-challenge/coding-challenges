# 🔄 Smart Team Rotator

A robust, TypeScript-based library and CLI for managing fair team rotations. This solution ensures fair distribution of duties while respecting member availability and avoiding immediate repetition.

## 🚀 Quick Start

### Prerequisites

- Node.js (v18+)
- npm

### Installation & Running

1. **Install dependencies:**

```bash
npm install

```

2. **Run Tests:**

```bash
npm test

```

3. **Run the example:**

```bash
npm run build
npm start

```

4. **Run the CLI:**

```bash
npx ts-node src/cli.ts
```

---

## 🏗️ Design Decisions

### 1. Design Pattern: **Iterator Pattern**

I chose the **Iterator Pattern** as the core architectural backbone.

- **Why:** The problem is fundamentally about traversing a collection (the team) in a specific order without exposing the underlying representation (the array indices).
- **Implementation:** The `TeamRotator` class implements a `RotationIterator` interface with standard `next()` and `hasNext()` methods.
- **Benefit:** This decouples the _traversal logic_ (skipping inactive members, handling wrap-arounds) from the _data structure_, making the code cleaner and easier to test.

### 2. Algorithm: Stateful Round-Robin

Instead of shuffling the array or using random selection, I implemented a deterministic **Round-Robin** approach with an index pointer.

- **Fairness:** Guarantees that every active member gets a turn before anyone goes twice.
- **State:** We track `currentIndex` and `lastSelectedId`.
- **Logic:** When `next()` is called, we iterate forward from `currentIndex`, skipping inactive members. If the candidate matches `lastSelectedId` (and other options exist), we skip them to prevent immediate repetition.

### 3. Immutable Data

The `TeamRotator` creates a defensive copy (`[...members]`) of the input array upon initialization. This prevents external mutations (e.g., modifying the original array elsewhere in the app) from corrupting the iterator's internal state index.

---

## ⚖️ Trade-offs

### 1. Time Complexity vs. Space Complexity

- **Decision:** I chose an **O(N)** traversal for `next()` calls rather than maintaining separate lists for "active" and "inactive" members.
- **Trade-off:** In the worst-case scenario (everyone inactive except one), the iterator must loop through the whole array.
- **Justification:** For a "Team" (typically <50 people), O(N) is negligible. This approach significantly reduces state synchronization bugs that occur when trying to maintain dual lists.

### 2. In-Memory State

- **Decision:** The state is held entirely in memory.
- **Trade-off:** If the application restarts, the rotation order resets to the beginning (or must be manually re-seeded via `init-with`).
- **Justification:** Database persistence was explicitly out of scope. This keeps the library lightweight and portable.

### 3. Conditional "No Immediate Repetition"

- **Decision:** This rule is **best-effort**. The system attempts to skip the previously selected member to ensure variety, but allows repetition if that member is the only one currently active.
- **Justification:** Availability takes precedence over variety. It is better to return the same person twice than to throw an error or return nothing when work needs to be done.

---

## 🔮 With More Time...

If the timeline were extended beyond 4-6 hours, I would improve the following:

1. **Persistence Layer:** Introduce a simple JSON file adapter or Redis connection to save the `currentIndex` and `lastSelectedId`, allowing the rotation to survive server restarts.
2. **Concurrency Support:** Although out of scope, a production API would need mutexes/locking around the `next()` method to prevent race conditions if two requests come in simultaneously.
3. **Weighted Rotation:** Add support for "seniority" or "capacity" weights, where some members might appear in the rotation more or less frequently than others.
4. **Strategy Pattern injection:** Allow passing different rotation strategies (e.g., `RandomWeightedStrategy`, `StrictRoundRobinStrategy`) into the rotator at runtime.

---

## 🤖 Session Notes

- **Coding:** Implementation logic and architecture generated using **Claude 3.5 Sonnet (Latest)**.
- **Documentation:** README generation and explanation of design decisions generated using **Gemini Pro**.
