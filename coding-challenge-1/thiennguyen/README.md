# Smart Team Rotator Library

A minimal, well-tested library for managing fair team rotations with support for active/inactive members and prevention of immediate repetition.

## 📋 Table of Contents

- [Overview](#overview)
- [Project Structure](#project-structure)
- [Tech Stack](#tech-stack)
- [Installation](#installation)
- [Usage](#usage)
- [Business Logic](#business-logic)
- [Design Patterns](#design-patterns)
- [Testing](#testing)
- [Examples](#examples)
- [Trade-offs](#trade-offs)

## 🎯 Overview

The Team Rotator library provides a simple API for rotating through team members with the following features:

- ✅ Round-robin rotation through active members
- ✅ No immediate repetition (same member won't be selected twice in a row)
- ✅ Automatic skipping of inactive members
- ✅ Fair distribution over time
- ✅ Support for getting next member or next N members
- ✅ Minimal state management (tracks only last selected member)

## 📁 Project Structure

```
coding-challenge-1/
├── packages/
│   └── team-rotator/          # Main library package
│       ├── src/
│       │   ├── index.ts       # Public API exports
│       │   ├── team-rotator.ts        # Main TeamRotator class
│       │   ├── member-iterator.ts    # Iterator pattern implementation
│       │   ├── types.ts               # Type definitions
│       │   ├── team-rotator.spec.ts   # Tests for TeamRotator
│       │   └── member-iterator.spec.ts # Tests for MemberIterator
│       ├── package.json
│       ├── project.json       # NX project configuration
│       ├── jest.config.ts     # Jest configuration
│       └── tsconfig.*.json    # TypeScript configurations
├── apps/
│   └── examples/              # Example application
│       ├── example1-basic-rotation.ts
│       ├── example2-no-repetition.ts
│       ├── example3-skip-inactive.ts
│       ├── example4-next-n-members.ts
│       ├── example5-single-active.ts
│       ├── example6-all-inactive.ts
│       ├── example7-add-new-member.ts
│       ├── tsconfig.json
│       ├── package.json
│       └── project.json       # NX project configuration
├── package.json              # Root package.json
├── nx.json                   # NX workspace configuration
├── tsconfig.json             # Root TypeScript configuration
└── README.md                 # This file
```

## 🛠️ Tech Stack

- **NX Monorepo** - Monorepo management and build system
- **Node.js LTS** - Runtime environment (managed via `.nvmrc`)
- **TypeScript** - Type-safe development
- **Jest** - Testing framework with >90% code coverage requirement
- **Yarn** - Package manager

### Third-party Libraries

- `nx` - Monorepo tooling
- `@nx/node` - NX Node.js plugin
- `@nx/jest` - NX Jest plugin
- `jest` - Testing framework
- `ts-jest` - TypeScript support for Jest
- `typescript` - TypeScript compiler

## 🚀 Installation

### Prerequisites

- **Node.js**: Version `>= 20.0.0` (specified in `.nvmrc` and `engines` field)
- **Yarn**: Version `>= 1.22.22` (specified in `packageManager` field)

### Setup

1. **Switch to the correct Node.js version (using nvm):**

   ```bash
   nvm use
   # or if nvm doesn't auto-detect:
   nvm install $(cat .nvmrc)
   nvm use $(cat .nvmrc)
   ```

2. **Verify versions:**

   ```bash
   node --version  # Should be v20.x.x or higher
   yarn --version  # Should be 1.22.22 or higher
   ```

3. **Install dependencies:**

   ```bash
   yarn install
   ```

4. **Build the library:**

   ```bash
   yarn build
   ```

5. **Run tests:**

   ```bash
   yarn test
   ```

6. **Check test coverage:**
   ```bash
   yarn test:coverage
   ```

## 💻 Usage

### Basic Example

```typescript
import { TeamRotator } from '@team-rotator/core';

const members = [
  { id: 1, name: 'Alice', isActive: true },
  { id: 2, name: 'Bob', isActive: true },
  { id: 3, name: 'Charlie', isActive: true },
];

const rotator = new TeamRotator(members);

// Get next member
const nextMember = rotator.getNext();
console.log(nextMember.name); // "Alice"

// Get next 2 members
const nextTwo = rotator.getNextN(2);
console.log(nextTwo.map((m) => m.name)); // ["Bob", "Charlie"]
```

### Handling Inactive Members

```typescript
const members = [
  { id: 1, name: 'Alice', isActive: true },
  { id: 2, name: 'Bob', isActive: false }, // inactive
  { id: 3, name: 'Charlie', isActive: true },
];

const rotator = new TeamRotator(members);
const next = rotator.getNext(); // Returns "Alice" (skips Bob)
```

### Manual Last Selected Member

```typescript
const rotator = new TeamRotator(members);

// If Alice was selected externally
rotator.setLastSelectedMember(1);

// Next call will skip Alice
const next = rotator.getNext(); // Returns "Bob"
```

## 🧠 Business Logic

### Rotation Algorithm

The rotation follows a **round-robin** approach with the following rules:

1. **Active Members Only**: Only members with `isActive: true` are considered
2. **No Immediate Repetition**: The last selected member is excluded from the next selection
3. **Fair Distribution**: Members are selected in order, ensuring equal distribution over time
4. **Edge Case Handling**:
   - If only one active member exists, repetition is allowed
   - If no active members exist, an error is thrown

### State Management

The library maintains minimal state:

- **Last Selected Member ID**: Tracks only the most recently selected member
- **Iterator Position**: Internal position in the rotation cycle

This minimal state approach keeps the solution simple and avoids over-engineering.

### Fairness Guarantee

Fairness is ensured through:

- **Round-robin ordering**: Members are selected in a consistent order
- **Equal distribution**: Over N rotations, each active member is selected approximately N/activeCount times
- **Position tracking**: The iterator maintains position across calls, ensuring continuity

## 🎨 Design Patterns

### Iterator Pattern

The library implements the **Iterator Pattern** through the `MemberIterator` class. This pattern is used because:

1. **Separation of Concerns**: The iteration logic is separated from the main `TeamRotator` class
2. **Encapsulation**: The internal mechanics of cycling through members are hidden
3. **Flexibility**: Easy to extend with different iteration strategies in the future
4. **Testability**: The iterator can be tested independently

**Implementation Details:**

- `MemberIterator` encapsulates the logic for cycling through active members
- It maintains the current position and handles skipping inactive/excluded members
- The `next()` and `nextN()` methods provide a clean interface for iteration

**Why This Pattern:**

- The problem inherently involves iterating through a collection (team members)
- The Iterator pattern provides a clean abstraction for this iteration
- It allows the main `TeamRotator` class to focus on business logic rather than iteration mechanics

## 🧪 Testing

The library has comprehensive test coverage (>90%) with tests covering:

- ✅ Basic rotation through all members
- ✅ No immediate repetition rule
- ✅ Skipping inactive members
- ✅ Getting next N members
- ✅ Edge cases (single active member, all inactive, empty team)
- ✅ Fair distribution over multiple rotations
- ✅ Manual state management

### Running Tests

```bash
# Run all tests
yarn test

# Run with coverage report
yarn test:coverage
```

### Test Coverage

The project maintains high code coverage:

- Statements: >90%
- Branches: >82% (some defensive branches are difficult to test in practice)
- Functions: >90%
- Lines: >90%

## 📚 Examples

The `apps/examples/` folder contains an example application demonstrating all scenarios:

1. **example1-basic-rotation.ts** - Simple round-robin rotation
2. **example2-no-repetition.ts** - Preventing immediate repetition
3. **example3-skip-inactive.ts** - Skipping inactive members
4. **example4-next-n-members.ts** - Getting multiple members at once
5. **example5-single-active.ts** - Edge case with one active member
6. **example6-all-inactive.ts** - Edge case with no active members

### Running Examples

```bash
# Build the library first
yarn build

# Run specific example using NX
nx run example:1
nx run example:2
# ... etc

# Or run all examples
nx run example:all

# Or directly with ts-node
npx ts-node -r tsconfig-paths/register apps/examples/example1-basic-rotation.ts
```

## ⚖️ Trade-offs

### Conscious Design Decisions

1. **Minimal State Management**
   - **Decision**: Only track last selected member, not full history
   - **Rationale**: Keeps the solution simple and avoids unnecessary complexity
   - **Trade-off**: Cannot provide full rotation history, but this was explicitly out of scope

2. **In-Memory Only**
   - **Decision**: No persistence layer
   - **Rationale**: Matches requirements (explicitly out of scope)
   - **Trade-off**: State is lost on application restart, but acceptable for this use case

3. **Single Rotation Strategy**
   - **Decision**: Only round-robin rotation implemented
   - **Rationale**: Requirements specify "one clean solution, not multiple strategies"
   - **Trade-off**: Less flexible, but simpler and more maintainable

4. **Iterator Pattern Over Strategy Pattern**
   - **Decision**: Used Iterator pattern instead of Strategy pattern
   - **Rationale**: The problem is fundamentally about iteration, not algorithm selection
   - **Trade-off**: Less extensible for multiple rotation algorithms, but more appropriate for the core problem

## 📝 License

ISC
