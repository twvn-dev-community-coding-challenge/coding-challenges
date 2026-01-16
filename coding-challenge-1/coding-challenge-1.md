# 🧩 Coding Challenge #1

## **Build a Smart Team Rotator API – v1**

> A minimal service that returns fair rotations for a team using simple rules.

---

## 🎯 Goal

Build a **small API or library** that:

- Maintains a team of members
- Returns the **next person** (or next N people) for a duty
- Ensures **no immediate repetition**
- Is easy to understand, test, and explain

This is **not** about completeness — it’s about **clarity and correctness**.

---

## ⏱️ Timebox

- **4–6 hours total**
- Aim for **one clean solution**, not multiple strategies

---

## 🛠️ Functional Requirements

### 1️⃣ Team & Members

- A team has a list of members
- A member has:
  - `id`
  - `name`
  - `isActive` (boolean)

👉 No CRUD API required

👉 Hard-coded or in-memory setup is acceptable

---

### 2️⃣ Rotation Logic

- API/function returns:
  - The **next member** (default)
  - Optionally the **next N members**
- Rules:
  - Must not return the **same member twice in a row**
  - Must **skip inactive members**
  - Rotation must be **fair over time** ("How will you ensure fairness over time, or a number of rotations?")

---

### 3️⃣ History

- Track **last selected member only**
- Full history is **NOT required**

📌 This keeps state management trivial.

---

### 4️⃣ Design Pattern Requirement

- Your solution **must apply at least one design pattern**
- Suggested patterns (choose what fits your design):
  - **Strategy Pattern** – for rotation algorithms
  - **Singleton Pattern** – for team state management
  - **Iterator Pattern** – for cycling through members
  - **Factory Pattern** – for creating rotation instances
  - **State Pattern** – for managing rotation state
  - Or any other relevant pattern that improves your design

📌 Be prepared to explain:
- Which pattern(s) you chose
- Why it was appropriate for this problem
- How it improved your solution

---

## 🚫 Explicitly Out of Scope

To avoid over-engineering:

- ❌ Database persistence
- ❌ Multiple rotation strategies
- ❌ Authentication
- ❌ UI
- ❌ Concurrency handling
- ❌ Distributed systems

---

## 📚 Concrete Examples

To help you understand the requirements, here are detailed text examples:

### Example 1: Basic Rotation

**Scenario**: Your team has 4 active members.

**Team Setup**:

- Alice (id: 1, active)
- Bob (id: 2, active)
- Charlie (id: 3, active)
- Diana (id: 4, active)

**Expected Behavior**:

- 1st call → Returns: **Alice**
- 2nd call → Returns: **Bob**
- 3rd call → Returns: **Charlie**
- 4th call → Returns: **Diana**
- 5th call → Returns: **Alice** (rotation restarts)
- 6th call → Returns: **Bob**

✅ This demonstrates simple round-robin rotation.

---

### Example 2: No Immediate Repetition

**Scenario**: What if Alice was just selected manually or externally?

**Team Setup**:

- Alice (id: 1, active) ← **last selected**
- Bob (id: 2, active)
- Charlie (id: 3, active)

**Expected Behavior**:

- 1st call → Returns: **Bob** (NOT Alice, because she was last)
- 2nd call → Returns: **Charlie**
- 3rd call → Returns: **Alice** (now OK to return her)
- 4th call → Returns: **Bob**

✅ This demonstrates the "no immediate repetition" rule.

---

### Example 3: Skipping Inactive Members

**Scenario**: Bob becomes inactive in the middle of rotation.

**Team Setup**:

- Alice (id: 1, active)
- Bob (id: 2, **inactive**) ← unavailable
- Charlie (id: 3, active)
- Diana (id: 4, active)

**Expected Behavior**:

- 1st call → Returns: **Alice**
- 2nd call → Returns: **Charlie** (skips Bob)
- 3rd call → Returns: **Diana**
- 4th call → Returns: **Alice** (skips Bob)
- 5th call → Returns: **Charlie**

✅ This demonstrates skipping inactive members automatically.

---

### Example 4: Requesting Next N Members

**Scenario**: You need 2 people for a task.

**Team Setup**:

- Alice (id: 1, active)
- Bob (id: 2, active)
- Charlie (id: 3, active)
- Diana (id: 4, active)

**Expected Behavior**:

- Call `getNext(n=2)` → Returns: **[Alice, Bob]**
- Call `getNext(n=2)` → Returns: **[Charlie, Diana]**
- Call `getNext(n=2)` → Returns: **[Alice, Bob]** (rotation restarts)

✅ This demonstrates returning multiple members at once.

---

### Example 5: Edge Case - Only One Active Member

**Scenario**: Everyone except Alice is inactive.

**Team Setup**:

- Alice (id: 1, active) ← only active member
- Bob (id: 2, **inactive**)
- Charlie (id: 3, **inactive**)

**Expected Behavior**:

- 1st call → Returns: **Alice**
- 2nd call → Returns: **Alice** (no choice, repetition is acceptable)
- 3rd call → Returns: **Alice**

✅ This demonstrates that the "no immediate repetition" rule is relaxed when there's only one active member.

---

### Example 6: Edge Case - All Members Inactive

**Scenario**: Everyone is on vacation or unavailable.

**Team Setup**:

- Alice (id: 1, **inactive**)
- Bob (id: 2, **inactive**)
- Charlie (id: 3, **inactive**)

**Expected Behavior**:

- Call `getNext()` → Returns: **Error or empty result**
- Suggested message: "No active members available"

✅ This demonstrates handling the case when no one is available.

---

## 🧪 Testing (Focused)

- Unit tests for:
  - No immediate repetition
  - Skipping inactive members
  - Rotation order correctness

👉 3–5 tests is enough.

---

## 📦 Deliverables

Each team submits:

- Git repository with **commit history** (frequent commits showing development progression)
- README (max 1 page) including:
  - How to run
  - Rotation approach
  - Design pattern(s) used and why
  - One trade-off they consciously made

📌 **Important**: If claiming TDD bonus points, your commit history should clearly show tests written before implementation code.

---

# 🏅 Knowledge Badges

This challenge awards the following **explicit skills**:

- 🟢 *API/Library Design Fundamentals*
- 🟢 *State Management Basics*
- 🟢 *Algorithm Design*
- 🟢 *Design Pattern Application*
- 🟢 *Unit Testing & Edge Cases*
- 🟢 *Test-Driven Development (TDD)*
- 🟢 *Clean Code Practices*
- 🟢 *Requirement Analysis*

---

# 🧮 Scoring Template

## 🏆 Total: **100 Points** (+ up to 20 bonus)

---

## 1️⃣ Core Logic Correctness — **45 pts**

| Criteria                | Points |
| ----------------------- | ------ |
| No immediate repetition | 10     |
| Skips inactive members  | 10     |
| Fair rotation           | 15     |
| Correct rotation order  | 10     |

---

## 2️⃣ Code Quality & Design — **25 pts**

| Criteria                          | Points |
| --------------------------------- | ------ |
| Design pattern implementation     | 10     |
| Readability & naming              | 5      |
| Separation of concerns            | 5      |
| Simplicity (no over-engineering)  | 5      |

---

## 3️⃣ Tests & Reliability — **15 pts**

| Criteria         | Points |
| ---------------- | ------ |
| Core logic tests | 10     |
| Edge cases       | 5      |

---

## 4️⃣ Communication & Demo — **15 pts**

| Criteria                      | Points |
| ----------------------------- | ------ |
| Clear explanation             | 5      |
| Design pattern justification  | 5      |
| README clarity                | 5      |

---

## 🌟 Bonus Points (Max +20)

| Bonus              | Points |
| ------------------ | ------ |
| TDD approach       | +10    |
| Junior-led demo    | +5     |
| Elegant simplicity | +5     |

📌 **Note on TDD Bonus**: To claim the TDD bonus points, your commit history must demonstrate that tests were written before implementation. Make sure to commit frequently throughout development to show your TDD workflow (red → green → refactor).
