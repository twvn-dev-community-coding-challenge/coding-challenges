# NoSleep - Coding Challenge 1

## Team Members

* **Trương Tuấn Lân** (Developer)
* **Thái Thị Ngọc Kim** (Developer)

## How to Run

1. **Prerequisites**: Ensure you have the [.NET 6.0 SDK](https://dotnet.microsoft.com/download) or later installed.
2. **Clone the project**:
```bash
git clone [your-repo-url]
cd TeamRotator

```


3. **Run Tests**:
To verify the rotation logic and edge cases, run:
```bash
dotnet test

```



## Approach

Our solution implements a **Round-Robin selection algorithm** designed for team member rotation.

* **Core Logic**: We use a `_lastSelectedId` tracker combined with modulo arithmetic `(index + i) % count`. This ensures that even if members are added or skipped, the rotation always moves forward smoothly.
* **State Management**: The `TeamRotator` class maintains internal state to remember the last person picked, allowing for consistent results across multiple calls.

## Design Decisions

* **Tech Stack**: We chose **C#** and **.NET** for its strong typing and excellent testing ecosystem.
* **Testing Framework**: We utilized **xUnit** and **FluentAssertions**. We chose FluentAssertions because it makes test expectations much more readable (e.g., `result.Should().Be("Bob")`).
* **Pattern**: We followed a **Stateful Service** pattern. By encapsulating the list and the "pointer" to the last selected member, the consumer doesn't need to worry about the internal rotation math.
* **Trade-offs**: We prioritized **simplicity and predictability** over complex randomized selection. This ensures that every active member gets an equal turn.

## Challenges Faced

* **Handling Inactive Members**: One challenge was ensuring that the rotation doesn't get "stuck" when several consecutive members are marked as inactive.
* **The Solution**: We implemented a `for` loop that iterates through the entire member list starting from the last index, skipping inactive entries until a valid candidate is found.

## What We Learned

* **Modulo Arithmetic**: Deepened our understanding of using `%` for circular data structures.
* **Edge Case Coverage**: Learned how to properly test for `InvalidOperationException` when no active members are available.

## With More Time, We Would...

* **Persistence**: Save the rotation state to a JSON file or database so it persists after the application restarts.
* **Weighted Rotation**: Allow some members to be picked more or less frequently based on a "weight" property.
* **UI**: Build a simple CLI or Web interface to visualize the rotation.

## AI Tools Used

* **Gemini**: Used for logic verification, brainstorming edge cases for the unit tests, and drafting this documentation.
