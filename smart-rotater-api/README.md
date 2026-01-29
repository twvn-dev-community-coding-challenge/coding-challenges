# Smart Team Rotator API

A fair and efficient team rotation service built with Node.js and TypeScript. This API ensures balanced member selection with no immediate repetition, following SOLID principles and design patterns.

## 🎯 Features

- ✅ **Fair Round-Robin Rotation**: Ensures each team member gets an equal turn
- ✅ **No Immediate Repetition**: Prevents the same member from being selected twice in a row
- ✅ **Inactive Member Handling**: Automatically skips inactive team members
- ✅ **Multiple Member Selection**: Request multiple members for group tasks
- ✅ **Singleton Pattern**: Ensures single team state throughout the application
- ✅ **Strategy Pattern**: Allows different rotation algorithms to be plugged in
- ✅ **Comprehensive Testing**: 10+ unit tests covering edge cases

## 📋 Requirements

- Node.js 16+
- npm or yarn

## 🚀 Quick Start

### Installation

```bash
npm install
```

### Running the Server

```bash
# Development mode with hot reload
npm run dev

# Production build and run
npm run build
npm start
```

The server will start on `http://localhost:3000`

### Running Tests

```bash
# Run all tests
npm test

# Run tests in watch mode
npm run test:watch
```

## 📚 API Endpoints

### Team Management

#### Get All Members
```
GET /api/team/members
```

Response:
```json
{
  "success": true,
  "data": [
    { "id": 1, "name": "Alice", "isActive": true },
    { "id": 2, "name": "Bob", "isActive": true }
  ],
  "count": 2
}
```

#### Get Active Members Only
```
GET /api/team/members/active
```

#### Add Member
```
POST /api/team/members
Content-Type: application/json

{
  "id": 5,
  "name": "Eve",
  "isActive": true
}
```

#### Update Member Status
```
PATCH /api/team/members/:id
Content-Type: application/json

{
  "isActive": false,
  "name": "Alice Updated"
}
```

### Rotation

#### Get Next Member(s)
```
GET /api/rotation/next?count=1
```

Query Parameters:
- `count`: Number of members to return (default: 1)

Response:
```json
{
  "success": true,
  "data": [
    { "id": 1, "name": "Alice", "isActive": true }
  ],
  "count": 1
}
```

#### Get Last Selected Member
```
GET /api/rotation/last
```

#### Manually Set Last Selected Member
```
POST /api/rotation/set-last
Content-Type: application/json

{
  "memberId": 2
}
```

## 🏗️ Architecture & Design Patterns

### 1. **Strategy Pattern** 📋
**File**: `src/strategies/RotationStrategy.ts`

The rotation algorithm is abstracted into a `RotationStrategy` interface. This allows:
- Different rotation algorithms to be implemented (Round-Robin, Random, Load-Based, etc.)
- Easy testing and switching strategies at runtime
- Separation of concerns between algorithm and team management

**Implementation**: `RoundRobinStrategy` provides fair, sequential rotation.

**Why**: Makes the code extensible and testable. Adding new rotation strategies requires only implementing the interface without modifying existing code.

### 2. **Singleton Pattern** 🔒
**File**: `src/managers/TeamManager.ts`

The `TeamManager` is implemented as a Singleton to ensure:
- Only one instance of team state exists throughout the application
- Consistent team data across all API requests
- Easy state management without global variables

**Why**: Prevents multiple instances of team state and ensures data consistency.

### 3. **Type Safety with TypeScript** 🎯
- Strong typing for all domain objects
- Interfaces for contracts (`Member`, `RotationResult`, `RotationStrategy`)
- Strict `tsconfig.json` for compile-time safety

## 🧪 Testing Coverage

### Test Cases Implemented

1. **Basic Rotation Order** - Verifies round-robin sequence
2. **No Immediate Repetition** - Ensures same member isn't selected twice
3. **Skip Inactive Members** - Confirms inactive members are skipped
4. **Multiple Member Selection** - Tests requesting N members
5. **Edge Case: Only One Active Member** - Handles repetition when necessary
6. **Edge Case: No Active Members** - Returns proper error
7. **Fairness Over Time** - Validates equal distribution over multiple rotations
8. **Member Management** - Tests adding/removing members
9. **Track Last Selection** - Verifies state persistence
10. **Active Member Filtering** - Confirms proper member status handling

### Running Tests

```bash
npm test                 # Run all tests once
npm run test:watch      # Run tests in watch mode
```

## 🔄 How It Works

### Basic Flow

1. **Initialize**: Create a `TeamManager` with a `RoundRobinStrategy`
2. **Add Team**: Add members to the team
3. **Rotate**: Call `getNext()` to get the next member
4. **Track**: The manager tracks the last selected member to prevent repetition

### Example Usage

```typescript
import { TeamManager } from './managers/TeamManager';
import { RoundRobinStrategy } from './strategies/RotationStrategy';

// Create manager with strategy
const strategy = new RoundRobinStrategy();
const manager = TeamManager.getInstance(strategy);

// Add team members
manager.addMembers([
  { id: 1, name: 'Alice', isActive: true },
  { id: 2, name: 'Bob', isActive: true },
  { id: 3, name: 'Charlie', isActive: true },
]);

// Get next member
const result = manager.getNext(1);
console.log(result.members[0].name); // Alice

// Get next 2 members
const result2 = manager.getNext(2);
console.log(result2.members.map(m => m.name)); // [Bob, Charlie]

// Skip inactive members
manager.updateMember(2, { isActive: false });
const result3 = manager.getNext(1);
console.log(result3.members[0].name); // Diana (Bob is skipped)
```

## 📊 Design Decisions & Trade-offs

### Decision 1: In-Memory State
**Choice**: Use in-memory `Map` for team storage instead of database

**Trade-off**:
- ✅ Simplicity and fast access
- ❌ Data lost on server restart
- ✅ Meets challenge requirements (no database required)

### Decision 2: Singleton for Team Manager
**Choice**: Implement `TeamManager` as Singleton

**Trade-off**:
- ✅ Ensures single source of truth for team state
- ✅ Simple to use across the application
- ❌ More difficult to test in some scenarios (mitigated with `reset()` method)

### Decision 3: Strategy Pattern for Rotation
**Choice**: Use Strategy pattern instead of hardcoding rotation logic

**Trade-off**:
- ✅ Easy to add new rotation algorithms later
- ✅ Better testability
- ❌ Slight over-engineering for current scope (but adds future flexibility)

### Decision 4: Track Only Last Member
**Choice**: Store only the last selected member ID instead of full history

**Trade-off**:
- ✅ Minimal memory footprint
- ✅ Fast state updates
- ❌ Cannot generate full history reports
- ✅ Sufficient for "no immediate repetition" requirement

## 🎓 What We Learned

1. **Design Patterns Matter**: Strategy and Singleton patterns made the code more maintainable and testable
2. **TypeScript Benefits**: Strong typing caught many potential bugs at compile time
3. **Clear Requirements**: Breaking down the challenge into concrete test cases helped avoid over-engineering
4. **Testing First Approach**: Writing tests first helped clarify the expected behavior

## ⏰ Time Investment

- Planning & Requirements Analysis: 30 minutes
- Design & Architecture: 45 minutes
- Core Implementation: 60 minutes
- Testing & Edge Cases: 60 minutes
- Documentation: 30 minutes

**Total: ~3.5 hours**

## 🚀 Future Enhancements (If More Time)

1. **Persistent Storage**: Add database layer with migrations
2. **Advanced Strategies**: Implement weighted rotation, preference-based selection
3. **Rotation History**: Store full rotation history for analytics
4. **Authentication & Authorization**: Add role-based access control
5. **Metrics & Monitoring**: Track rotation fairness metrics
6. **API Validation**: Add Joi or Zod for request validation
7. **Rate Limiting**: Prevent abuse of rotation API
8. **Logging**: Add structured logging for debugging
9. **Docker Support**: Containerize for easy deployment
10. **GraphQL Alternative**: Offer GraphQL API alongside REST

## 📝 Commit History

The git history shows:
1. Initial project setup (package.json, tsconfig.json)
2. Type definitions and interfaces
3. Strategy pattern implementation
4. Team Manager implementation
5. API routes
6. Comprehensive tests
7. Documentation and README

This progression demonstrates the development workflow and careful planning.

## 📧 Support

For questions or issues, please refer to the coding challenge guidelines or contact the team.

---

**Built with ❤️ for the TWVN Coding Challenge 2026**
