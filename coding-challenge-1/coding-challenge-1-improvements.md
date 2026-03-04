# Coding Challenge #1 — Improvement Feedback for All Teams

> This document consolidates all improvement areas identified across every submission. Each issue is supported by specific evidence (file, line, and/or code). Use this as a learning reference after the challenge concludes.

---

## Team: Thien Nguyen (PR #2)

**Tech Stack:** TypeScript, NX Monorepo, Jest

### 1. Position Tracking Bug — Unguarded `-1` Return

**Location:** `member-iterator.ts:113-118`

**Issue:** `findCurrentPositionInActive()` returns `-1` when the member at `currentIndex` is inactive. This value is then used as an array index — `activeMembers[-1]` evaluates to `undefined` — causing a `TypeError` crash on the first `next()` call when `members[0]` is inactive and 2+ active members exist.

```typescript
// BUG: returns -1 when current member is not in the active list
private findCurrentPositionInActive(activeMembers: Member[]): number {
    const currentMember = this.members[this.currentIndex];
    let position = activeMembers.findIndex((m) => m.id === currentMember.id);
    return position; // -1 if not found — NOT GUARDED
}
```

**Fix:** Default to `0` when the position is not found:

```typescript
return position >= 0 ? position : 0;
```

---

### 2. Dependency Inversion Violation — Hardcoded Iterator Instantiation

**Location:** `team-rotator.ts:30`

**Issue:** `TeamRotator` directly instantiates `new MemberIterator(this.members)` in its constructor. This creates tight coupling — you cannot inject a mock iterator for isolated unit testing, nor swap the iteration strategy without modifying `TeamRotator`.

```typescript
constructor(members: Member[]) {
    this.members = [...members];
    this.iterator = new MemberIterator(this.members); // hardcoded — no injection point
}
```

**Fix:** Define an interface and inject through the constructor:

```typescript
interface IMemberIterator {
    next(excludeId?: number): Member | null;
    nextN(count: number, excludeId?: number): Member[];
}

constructor(members: Member[], iterator?: IMemberIterator) {
    this.members = [...members];
    this.iterator = iterator ?? new MemberIterator(this.members);
}
```

---

### 3. False-Pass Risk in Error Assertion Tests

**Location:** `member-iterator.spec.ts:23-30`

**Issue:** Error-throwing tests use `try/catch` blocks. If the constructor does **not** throw (e.g., due to a regression), the catch block is never entered, no assertions run, and the test silently passes — a false positive.

```typescript
// If new MemberIterator(members) stops throwing, this test still passes silently
try {
    new MemberIterator(members);
} catch (err) {
    expect(err instanceof NoActiveMembersError).toBeTruthy();
}
```

**Fix:** Use Jest's `expect().toThrow()`:

```typescript
expect(() => new MemberIterator(members)).toThrow(NoActiveMembersError);
expect(() => new MemberIterator(members)).toThrow('No active members available');
```

---

### 4. Missing Test: First Member Inactive

**Issue:** Every test places an active member at index `0`. No test covers `[inactive, active, active]` at construction — the exact scenario that triggers the `-1` position bug above. If this test had existed, the crash would have been caught immediately.

**Fix:** Add an explicit test:

```typescript
it('should handle first member being inactive', () => {
    const members: Member[] = [
        { id: 1, name: 'Alice', isActive: false },
        { id: 2, name: 'Bob', isActive: true },
        { id: 3, name: 'Charlie', isActive: true },
    ];
    const iterator = new MemberIterator(members);
    expect(iterator.next()?.name).toBe('Bob');
    expect(iterator.next()?.name).toBe('Charlie');
});
```

---

### 5. Redundant Conditional in `getNextN`

**Location:** `team-rotator.ts:65-67`

**Issue:** After throwing when `members.length === 0`, the next line checks `if (members.length > 0)` — which is always true. Dead condition adds noise.

```typescript
if (members.length === 0) throw new NoActiveMembersError();
if (members.length > 0) {              // always true
    this.lastSelectedMemberId = ...;
}
```

**Fix:** Remove the redundant guard.

---

### 6. No TDD Evidence in Commit History

**Issue:** The initial commit (`c4bfa5a`) contained the complete solution in one go — a "big-bang" commit. A subsequent commit was `fix(test): fixed failed tests`, indicating tests were adjusted *after* implementation. There is no observable Red-Green-Refactor cycle.

**Fix:** Adopt granular, feature-by-feature commits:
```
test: add failing test for basic round-robin
feat: implement basic MemberIterator.next()
test: add failing test for no-repetition rule
feat: add excludeId support to MemberIterator.next()
refactor: extract getActiveMembers() helper
```

---

## Team: NHA (PR #6)

**Tech Stack:** Kotlin, Gradle, JUnit 5

### 1. No Design Pattern Named or Implemented

**Location:** `TeamRotator.kt` (entire file), `README.md`

**Issue:** The challenge explicitly requires a design pattern. NHA's PR description states "TDD instead of using complex data structures" — but TDD is a *methodology*, not a design pattern. No interface is defined, no pattern is applied (Iterator, Strategy, State, etc.), and the README makes no mention of any pattern.

**Fix:** Define an interface for `TeamRotator`:

```kotlin
interface ITeamRotator {
    fun rotate(): Member
    fun rotateNMembers(n: Int): List<Member>
    fun markMemberInactive(name: String)
    fun markMemberActive(name: String)
    fun isMemberActive(name: String): Boolean
}

class TeamRotator(members: List<Member>) : ITeamRotator { ... }
```

Then justify the choice in the README: "We applied a Stateful Service pattern..."

---

### 2. Dead Code Left in Test Directory

**Location:** `src/test/kotlin/Member.kt`, `src/test/kotlin/TeamRotator.kt`

**Issue:** These are leftover TDD scaffolding stubs that were superseded when code moved to `src/main/kotlin`. They now conflict with — and shadow — the production classes.

- `Member.kt` (test): `data class Member(val fullName: String)` — 3 lines, no package, duplicate
- `TeamRotator.kt` (test): 20-line stub with `rotateCount: Int = 0` — entirely replaced

**Fix:** Delete both files. They should have been removed during the refactoring step.

---

### 3. IntelliJ IDEA Template Committed as `Main.kt`

**Location:** `src/main/kotlin/Main.kt`

**Issue:** This is the default IntelliJ new-project template, complete with auto-generated TIP comments:

```kotlin
//TIP To <b>Run</b> code, press <shortcut actionId="Run"/> or
fun main() {
    val name = "Kotlin"
    println("Hello, " + name + "!")
```

This has no business being in a final submission. It shows the development environment wasn't cleaned up before submission.

**Fix:** Delete `Main.kt` (or replace with a real usage demo). Add `*.iml` to `.gitignore` to prevent IDE config files (`solution.iml`) from being committed.

---

### 4. Inconsistent Semicolons in Kotlin Code

**Location:** `TeamRotator.kt` lines 6, 25, 39, 52, 53

**Issue:** Kotlin's convention is no semicolons, but the code is inconsistent — some lines have them, others don't:

```kotlin
private var lastSelectedIndex: Int = -1;   // semicolon
fun rotate(): Member {
    throwErrorIfAllMemberInactive()         // no semicolon
    rotateLastSelectedIndex()              // no semicolon
    ...
    return membersList[lastSelectedIndex]   // no semicolon
}
fun lastSelectedMember(): Member {
    return membersList[lastSelectedIndex];  // semicolon
}
```

**Fix:** Remove all semicolons. Configure `ktlint` or run IntelliJ's Kotlin formatter (`Code → Reformat Code`).

---

### 5. README Missing Algorithm Explanation and Pattern Discussion

**Issue:** The README explains how to build and run, and mentions the Kotlin language choice, but is silent on:
- How the rotation algorithm actually works
- What design pattern was applied (or a justification for not using one)
- Time and space complexity

**Fix:** Add a section:

```markdown
### Algorithm: Index-Based Round-Robin

1. Maintain `lastSelectedIndex` starting at -1
2. On each `rotate()`, increment the index (wrapping at list end)
3. Skip inactive members by continuing to increment
4. Pre-check for any active member to prevent infinite loops

**Complexity:** O(n) time per rotation, O(1) space
```

---

### 6. `data class Member` with Mutable State

**Location:** `Member.kt:3`

**Issue:** Kotlin's `data class` implies value semantics — `equals()`, `hashCode()`, and `copy()` are generated based on constructor parameters. Using `private var isActive` in a `data class` means two members with the same name but different active status are considered equal, and `copy()` cannot clone active status. This is a Kotlin anti-pattern.

```kotlin
data class Member(val fullName: String, private var isActive: Boolean = true)
```

**Fix:** Use a regular class with explicit `equals`/`hashCode`:

```kotlin
class Member(val fullName: String, isActive: Boolean = true) {
    var isActive: Boolean = isActive
        private set

    override fun equals(other: Any?) = other is Member && fullName == other.fullName
    override fun hashCode() = fullName.hashCode()
}
```

---

### 7. No Test for Dynamic Status Changes Mid-Rotation

**Issue:** No test deactivates or reactivates a member *between* successive `rotate()` calls. While the algorithm handles this correctly (the `while` loop re-checks `isActive()` on every call), the behaviour is unverified.

**Fix:**

```kotlin
@Test
fun `rotation should skip member deactivated mid-rotation`() {
    val rotator = TeamRotator("Alice", "Bob", "Charlie")
    assertEquals(Member("Alice"), rotator.rotate())
    rotator.markMemberInactiveByName("Bob")
    assertEquals(Member("Charlie"), rotator.rotate()) // skips Bob
    assertEquals(Member("Alice"), rotator.rotate())   // wraps, skips Bob
}
```

---

## Team: BSOD (PR #5)

**Tech Stack:** TypeScript, Node.js, Jest

### 1. Unused `Team` Class (Dead Code)

**Location:** `src/models/team.ts`, `src/models/team.test.ts`

**Issue:** `Team` is a fully implemented class with `getActiveMembers()`, `getMemberById()`, `setLastSelectedMemberId()`, and its own 2-test file. However, `TeamRotator` never imports `Team` — it manages members directly. This is 55 lines of dead code that creates a misleading impression of the architecture.

```typescript
// team.ts — never imported by TeamRotator.ts
export default class Team {
    getActiveMembers(): Member[] { ... }
    getMemberById(memberId: MemberId): Member | undefined { ... }
    setLastSelectedMemberId(memberId: MemberId | null): void { ... }
}
```

**Fix:** Either delete `team.ts` and `team.test.ts`, or refactor `TeamRotator` to use `Team` as an internal data manager.

---

### 2. Shallow Copy Contradicts "Immutable State" Claim

**Location:** `TeamRotator.ts` constructor

**Issue:** The README claims "defensive copy prevents external mutations." But `[...members]` is a **shallow copy** — member objects are shared references. External code can still mutate member properties:

```typescript
const team = [createMember(1, 'Alice')];
const rotator = new TeamRotator(team);
team[0].isActive = false; // mutates the member inside the rotator too
```

The code comment "Member status is fixed at initialization" is also misleading — status *can* change via these shared references.

**Fix:** Deep-copy member objects:

```typescript
this.members = members.map(m => ({ ...m }));
```

Or use a `readonly` Member type that TypeScript enforces.

---

### 3. Missing Edge Case Tests

**Issue:** Tests cover the 6 challenge examples exactly, but stop there. Missing:

- Empty array constructor validation (code throws, but no test verifies it)
- `nextN(0)` — code returns `[]`, but no test
- `nextN(1)` — boundary
- Dynamic status changes: deactivate a member mid-rotation, verify skip

**Fix (example):**

```typescript
it('should throw when initialized with empty array', () => {
    expect(() => new TeamRotator([])).toThrow('Team must have at least one member');
});

it('should return empty array for nextN(0)', () => {
    const rotator = new TeamRotator([createMember(1, 'Alice')]);
    expect(rotator.nextN(0)).toEqual([]);
});
```

---

### 4. Challenge Specification Duplicated in Submission

**Location:** `coding-challenge-1/BSOD/coding-challenge-1.md` (325 lines)

**Issue:** The full challenge specification was copied into the submission directory. It already exists at the repo root. This adds 325 lines of non-functional content that obscures the actual solution.

**Fix:** Delete the file. Reference the original if needed: `See [challenge spec](../../coding-challenge-1.md)`.

---

### 5. `nextN()` Throws When `count > activeCount` — No Cyclic Option

**Location:** `TeamRotator.ts` — `nextN()` method

**Issue:** `nextN(5)` with 2 active members throws an error. The challenge implies cyclic rotation — requesting more members than are available should wrap around. Some use cases legitimately need `getNext(4)` from a 3-member team (scheduling weeks of rotation).

```typescript
if (count > activeCount) {
    throw new Error(`Cannot select ${count} members: only ${activeCount} active members available`);
}
```

**Fix:** Add an `allowRepetition` option:

```typescript
nextN(count: number, allowRepetition = false): Member[] {
    if (!allowRepetition && count > activeCount) {
        throw new Error(...);
    }
    // collect via repeated next() calls
}
```

---

### 6. Development History Lost — Commit Migration Strategy

**Issue:** The submitted branch contains only 1 team commit (`fix: clone source form BSOD branch`) because files were bulk-copied from the original branch to resolve a Git case-sensitivity conflict. All incremental development history was lost.

**Fix:** When migrating a branch, preserve history using `git cherry-pick` or `git rebase` rather than copying files manually:

```bash
git checkout -b coding-challenge-1/BSOD-2 main
git cherry-pick <first-commit>..<last-commit>
```

This preserves every commit — including TDD evidence — in the new branch.

---

## Team: NoSleep (PR #7)

**Tech Stack:** C# .NET 6.0, xUnit, FluentAssertions

### 1. Only 5 Tests — Thin Edge Case Coverage

**Issue:** The test file covers 5 scenarios, which mirrors the happy path and a couple of edge cases. A competitive submission should have 15–20+ tests. Specifically missing:

- Single active member (repetition allowed)
- Empty team (no members added)
- `GetNext(n)` where `n > active members` — cyclic wrapping
- `GetNext(0)` and `GetNext(-1)` — boundary inputs
- Duplicate member IDs added via `AddMember`
- `SetManualLastSelected(id)` with an ID that doesn't exist
- Quantitative fairness verification (e.g., 9 rotations / 3 members = exactly 3 each)

**Fix (examples):**

```csharp
[Fact]
public void GetNext_SingleActiveMember_AllowsRepetition()
{
    _rotator.AddMember(1, "Alice");
    _rotator.AddMember(2, "Bob", isActive: false);

    var first = _rotator.GetNext(1).First().Name;
    var second = _rotator.GetNext(1).First().Name;

    first.Should().Be("Alice");
    second.Should().Be("Alice", "repetition is acceptable when only one active member");
}

[Fact]
public void GetNext_FairDistribution_EachMemberSelectedEqually()
{
    _rotator.AddMember(1, "Alice");
    _rotator.AddMember(2, "Bob");
    _rotator.AddMember(3, "Charlie");

    var selections = Enumerable.Range(0, 9)
        .Select(_ => _rotator.GetNext(1).First().Name)
        .ToList();

    selections.Count(n => n == "Alice").Should().Be(3);
    selections.Count(n => n == "Bob").Should().Be(3);
    selections.Count(n => n == "Charlie").Should().Be(3);
}
```

---

### 2. No Formal GoF Design Pattern

**Issue:** The README names "Stateful Service" as the pattern — but this is an architectural description, not a recognized GoF pattern. The code has no interface (`ITeamRotator`), which prevents dependency injection, mocking in tests, and future extensibility.

**Fix:** Add an interface:

```csharp
public interface ITeamRotator
{
    void AddMember(int id, string name, bool isActive = true);
    List<Member> GetNext(int count = 1);
    void SetManualLastSelected(int id);
}

public class TeamRotator : ITeamRotator { ... }
```

Name the pattern in the README — "Iterator" or "Stateful Iterator" — and explain why you chose it.

---

### 3. Typo in Test File Name

**Location:** `TeamRotater.Tests.cs` (should be `TeamRotator.Tests.cs`)

**Issue:** "Rotater" vs "Rotator" — a single missing letter. Minor, but visible during a code review and inconsistent with the production class name `TeamRotator`.

**Fix:** Rename the file to `TeamRotator.Tests.cs`.

---

### 4. No Input Validation

**Issue:** Several inputs are accepted silently without validation:
- `AddMember` allows duplicate IDs — the second member with the same ID is simply added to the list
- `GetNext(0)` and `GetNext(-1)` return an empty list without error
- `SetManualLastSelected(999)` — ID doesn't exist, silently starts rotation from index 0

**Fix:**

```csharp
public void AddMember(int id, string name, bool isActive = true)
{
    if (_members.Any(m => m.Id == id))
        throw new ArgumentException($"Member with id {id} already exists");
    _members.Add(new Member(id, name, isActive));
}

public List<Member> GetNext(int count = 1)
{
    if (count < 0) throw new ArgumentOutOfRangeException(nameof(count), "Count must be >= 0");
    // ...
}
```

---

### 5. `Member.IsActive` Is Immutable — No Runtime Status Changes

**Location:** `Member.cs`

**Issue:** `IsActive` is a get-only property (`public bool IsActive { get; }`). While immutability is a strength, this means the "skip inactive members" feature only works for members configured as inactive at construction time. Deactivating a member at runtime — which the challenge implies — is not possible.

**Fix:** Consider allowing runtime status updates either directly on `Member` or via a `TeamRotator` method:

```csharp
public class TeamRotator
{
    public void SetMemberActive(int id, bool isActive)
    {
        var index = _members.FindIndex(m => m.Id == id);
        if (index == -1) throw new InvalidOperationException($"Member {id} not found");
        _members[index] = new Member(id, _members[index].Name, isActive);
    }
}
```

---

### 6. No XML Documentation on Public API

**Issue:** The public API (`AddMember`, `GetNext`, `SetManualLastSelected`) has no XML documentation comments. While the naming is self-documenting, XML docs enable IntelliSense tooltips for consumers of the class.

**Fix:**

```csharp
/// <summary>
/// Gets the next <paramref name="count"/> members in fair rotation order,
/// skipping inactive members and avoiding immediate repetition.
/// </summary>
/// <param name="count">Number of members to return. Defaults to 1.</param>
/// <exception cref="InvalidOperationException">Thrown when no active members available.</exception>
public List<Member> GetNext(int count = 1) { ... }
```

---

## Team: Eric (PR #3)

**Tech Stack:** TypeScript, Express.js, Jest

### 1. No TDD Evidence in Commit History

**Issue:** All 5 commits are implementation-first. The initial commit contains the full working solution. There is no visible Red-Green-Refactor cycle, no test-first commits.

```
commit 1: feat(coding-challenge-1): init commit  ← full solution in one commit
commit 2: feat(rotator): handled case add new member, assume TDD fails
commit 3: fix(test): fixed failed tests           ← tests fixed AFTER implementation
```

**Fix:** Practice granular TDD commits:

```
test: add failing test for basic round-robin rotation
feat: implement basic rotation
test: add failing test for no-repetition rule
feat: add lastMemberIndex tracking to prevent repetition
refactor: extract getActiveMembers() helper
```

---

### 2. Non-Null Assertion on Potentially Null Value

**Location:** `RotationService.ts:L88`

**Issue:** `selectedMembers?.members.length!` combines optional chaining (`?.`) with a non-null assertion (`!`). When `selectedMembers` is `null`, `selectedMembers?.members.length` evaluates to `undefined`, and `count > undefined` returns `false` in JavaScript — silently skipping the validation check entirely.

```typescript
// UNSAFE: null check comes AFTER the assertion
if (count > selectedMembers?.members.length!) {  // undefined! is still undefined
    return { success: false, message: 'Not enough active members...' };
}
if (selectedMembers === null) {  // too late — the above already evaluated with undefined
    return { success: false, message: 'No active members available' };
}
```

**Fix:** Reverse the check order:

```typescript
if (selectedMembers === null) {
    return { success: false, message: 'No active members available' };
}
if (count > selectedMembers.members.length) {
    return { success: false, message: `Not enough active members...` };
}
```

---

### 3. Incomplete Singleton Pattern — No Private Constructor

**Location:** `RotationService.ts`

**Issue:** `RotationService.getInstance()` exists, but the constructor is public. Any code (including tests) can call `new RotationService()`, bypassing the Singleton entirely. The tests themselves do this, proving the pattern is unenforced.

**Fix:**

```typescript
export class RotationService {
    private static instance: RotationService;

    private constructor() {}  // ← enforce Singleton

    static getInstance(): RotationService {
        if (!RotationService.instance) {
            RotationService.instance = new RotationService();
        }
        return RotationService.instance;
    }
}
```

---

### 4. Immutability Risk — `getMemberlist()` Returns Direct Reference

**Location:** `members.ts:L14`

**Issue:** `getMemberlist()` returns a direct reference to `DEFAULT_MEMBERS`. Any caller can push, splice, or sort the array and silently corrupt the rotation state.

```typescript
export function getMemberlist(): Member[] {
    return DEFAULT_MEMBERS; // direct reference — no copy
}
```

**Fix:**

```typescript
export function getMemberlist(): readonly Member[] {
    return [...DEFAULT_MEMBERS]; // defensive copy
}
```

---

### 5. Controller-Factory Coupling — DIP Violation

**Location:** `RotationController.ts:L43`

**Issue:** The controller directly calls `RotationStrategyFactory.getRoundRobinStrategy()`. This means the controller has knowledge of the Factory pattern internals, violating DIP. If you need to change the strategy selection logic, you must modify the controller.

**Fix:** Inject the strategy via the service or via dependency injection at a higher level, keeping the controller unaware of factory internals:

```typescript
// Inject at application startup
const strategy = RotationStrategyFactory.getRoundRobinStrategy();
const service = RotationService.getInstance(strategy);
```

---

### 6. Missing Edge Case Tests

**Issue:** No test for:
- Single active member (the no-repetition rule should be relaxed)
- Empty team initialization
- Dynamic active/inactive changes mid-rotation

**Fix (example):**

```typescript
it('should return the same member repeatedly when only one is active', async () => {
    // Set all members inactive except one
    // Call next 3 times — all should return the same member
});
```

---

## Team: Hoa Nguyen (PR #1)

**Tech Stack:** Python 3, pytest

### 1. Critical Bug: `active_count` Always Incremented in `add_member`

**Location:** `member_manager.py:29`

**Issue:** `self.active_count += 1` runs unconditionally, regardless of the member's actual status. If an inactive member is added, `active_count` is over-counted. Later, `get_next()` uses `min(n, active_count)` to decide how many iterations to run — but with an inflated count and no truly active members to find, the loop **runs forever**.

```python
def add_member(self, member_info: Member) -> None:
    # ...
    self.active_count += 1  # BUG: always increments, even if member is inactive
```

**Fix:**

```python
if member_info.status == Status.isActive:
    self.active_count += 1
```

---

### 2. Debug `print()` in Production Code

**Location:** `member_manager.py:23`

**Issue:** `print("Current members: ", self.members)` runs every time a member is added. This pollutes stdout in production and clearly indicates the code was not cleaned up before submission.

**Fix:** Delete the line.

---

### 3. Dead Code: `prev_id` Set But Never Read

**Location:** `member_manager.py:13, 77`

**Issue:** `self.prev_id` is assigned in `get_next()` but never read in any conditional, return, or assertion anywhere in the codebase. This is a strong signal that the developer intended an explicit no-repetition check but never completed it.

```python
self.prev_id = member_id  # set on every selection, never used
```

**Fix:** Either implement the intended check, or remove both the declaration and assignment.

---

### 4. Commented-Out Code in Final Submission

**Location:** `member_manager.py:64-68`

**Issue:** Three commented-out lines for handling `active_count == 0` and `active_count == 1` were left in the final code:

```python
# if self.active_count == 0:
#     return []
# ...
```

Final submissions should not contain commented-out code. Either implement it or remove it.

**Fix:** Delete the commented lines. Add tests instead to verify the all-inactive behaviour.

---

### 5. No Formal Design Pattern

**Issue:** The challenge requires at least one design pattern. `MemberManager` is a single monolithic class mixing member storage, status management, counter tracking, and rotation logic. There is no separation, no abstraction, and no mention of a pattern in the README.

**Fix:** Extract the rotation logic behind an Iterator interface:

```python
from abc import ABC, abstractmethod

class RotationIterator(ABC):
    @abstractmethod
    def next(self, n: int = 1) -> list:
        pass

class RoundRobinRotator(RotationIterator):
    def __init__(self, members: list):
        self._members = members
        self._index = 0

    def next(self, n: int = 1) -> list:
        active = [m for m in self._members if m.status == Status.ACTIVE]
        if not active:
            return []
        result = []
        for _ in range(n):
            result.append(active[self._index % len(active)])
            self._index += 1
        return result
```

---

### 6. Input Mutation in `add_member`

**Location:** `member_manager.py:27`

**Issue:** `member_info.id = self.member_count + 1` silently mutates the *caller's* `Member` object. The caller does not expect this:

```python
alice = Member(name="Alice", status=Status.isActive)
manager.add_member(alice)
print(alice.id)  # was None, now 1 — surprise mutation
```

**Fix:** Create a new `Member` internally rather than modifying the passed-in object:

```python
def add_member(self, member_info: Member) -> None:
    new_id = self.member_count + 1
    new_member = Member(name=member_info.name, status=member_info.status, id=new_id)
    self.members[new_id] = new_member
    self.member_count += 1
    if member_info.status == Status.isActive:
        self.active_count += 1
```

---

### 7. Wrong Argument Order in Test

**Location:** `test_member_manager.py:60`

**Issue:** The `Member` dataclass is defined as `Member(name, status, id)`. The test passes:

```python
Member(1, "Alice", Status.isActive)  # name=1, status="Alice", id=Status.isActive
```

This works at runtime because Python doesn't enforce dataclass field types, but the values are semantically in the wrong positions. If type validation were added, this test would fail.

**Fix:**

```python
Member(name="Alice", status=Status.isActive, id=1)
# or use the factory if one exists
```

---

### 8. Poor Variable Naming

**Issue:** Several variable names don't communicate intent:

| Name | Location | Problem | Better Name |
|------|----------|---------|-------------|
| `ans` | `member_manager.py:72` | Generic result variable | `selected_members` |
| `i` | `member_manager.py:71` | While-loop decrementor | `remaining` |
| `counter` | `member_manager.py` | Counter of what? | `rotation_counter` |
| `prev_id` | `member_manager.py` | Never used, unclear intent | (delete) |

---

### 9. Module-Scoped Shared Fixture — Tests Are Order-Dependent

**Location:** `test_member_manager.py` — `member_manager_object` fixture

**Issue:** A `module`-scoped fixture creates one `MemberManager` instance shared across all tests. Tests that mutate this shared state (adding members, rotating) leave residual state for subsequent tests, making test results dependent on execution order.

**Fix:** Use `function` scope (the default in pytest) so each test gets a fresh instance:

```python
@pytest.fixture  # default scope is "function"
def member_manager():
    return MemberManager()
```

---

### 10. Non-Pythonic Enum Values

**Location:** `status.py`

**Issue:** The `Status` enum uses camelCase values (`isActive`, `isNotActive`), which goes against Python's naming conventions. PEP 8 recommends uppercase for enum members.

```python
class Status(Enum):
    isActive = "isActive"       # non-Pythonic
    isNotActive = "isNotActive"
```

**Fix:**

```python
class Status(Enum):
    ACTIVE = "active"
    INACTIVE = "inactive"
```
