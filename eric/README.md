# Smart Team Rotator API

A fair and efficient team rotation service built with Node.js and TypeScript. This API ensures balanced member selection with no immediate repetition, following SOLID principles and design patterns.


## 🎯 Features

- ✅ **Fair Round-Robin Rotation**: Ensures each team member gets an equal turn
- ✅ **No Immediate Repetition**: Prevents the same member from being selected twice in a row
- ✅ **Inactive Member Handling**: Automatically skips inactive team members
- ✅ **Multiple Member Selection**: Request multiple members for group tasks
- ✅ **Singleton Pattern**: Ensures single service state throughout the application
- ✅ **Strategy Pattern**: Rotation algorithm abstracted for extensibility
- ✅ **Factory Pattern**: Centralized strategy instance creation
- ✅ **Index and ID Cache**: Caches both index and ID to optimize performance
- ✅ **Index-ID Sync Recovery**: Handles edge cases when member index and ID are out of sync

## 📋 Requirements

- Node.js 16+
- yarn
- Docker (optional)

## 🚀 Quick Start

### Using Docker (Recommended)
Requires: Install Docker first

```bash
cd eric
docker compose up
```

The server will start on `http://localhost:1902`

Note: You can update DEFAULT_MEMBERS to test without restarting Docker

### Local Development

```bash
# Install dependencies
yarn install

# Development mode with hot reload
yarn dev

# Production build and run
yarn build
yarn start
```

### Running Tests

```bash
# Run all tests
yarn test

# Run tests in watch mode
yarn test:watch
```

## 📚 API Endpoint

### ⭐ Core Rotation Endpoint

```
GET /api/next?count=1
```

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `count` | number | 1 | Number of members to return |

**Success Response (200):**
```json
{
  "success": true,
  "data": [
    { "id": 1, "name": "Alice" },
    { "id": 2, "name": "Bob" }
  ]
}
```

**Error Responses (400):**
```json
{ "success": false, "message": "Count must be at least 1" }
```
```json
{ "success": false, "message": "No active members available" }
```
```json
{ "success": false, "message": "Count cannot exceed total number of members (4)" }
```
```json
{ "success": false, "message": "Not enough active members to fulfill the request. Requested: 3, Available: 2" }
```

### Health Check
```
GET /api/health
```

**Response:**
```json
{ "status": "healthy" }
```

### Example Usage

```bash
# Get next 1 member (default)
curl http://localhost:1902/api/next

# Get next 2 members
curl http://localhost:1902/api/next?count=2

# Get next 4 members
curl http://localhost:1902/api/next?count=4
```

### 📮 Postman Collection

A Postman collection is available for easy API testing:

**Location:** `postman/coding-challenge-1.postman_collection.json`

**How to use:**
1. Open Postman
2. Click **Import** button
3. Select the file `postman/coding-challenge-1.postman_collection.json`
4. The collection includes:
   - `health-check` - GET `/api/health`
   - `get-next` - GET `/api/next?count=1`

## 🏗️ Architecture & Design Patterns

### Project Structure

```
eric/
├── src/
│   ├── index.ts                  # Express app entry point
│   ├── controllers/
│   │   └── RotationController.ts # HTTP request handling
│   ├── services/
│   │   └── RotationService.ts    # Business logic (Singleton)
│   ├── strategies/
│   │   └── RotationStrategy.ts   # Strategy + Factory patterns
│   ├── data/
│   │   └── members.ts            # Centralized member data
│   ├── routes/
│   │   └── api.ts                # Route definitions
│   ├── utils/
│   │   └── types.ts              # TypeScript interfaces
│   └── __tests__/
│       └── component.test.ts     # Unit tests
├── package.json
├── tsconfig.json
├── jest.config.js
├── Dockerfile
└── docker-compose.yml
```

### 1. **Strategy Pattern** 📋
**File**: `src/strategies/RotationStrategy.ts`

The rotation algorithm is abstracted into a `RotationStrategy` interface:

```typescript
export interface RotationStrategy {
    getNext(
        members: Member[], 
        lastMemberIndex: number | null, 
        lastMemberId: number | null, 
        count: number
    ): { 
        lastMemberIndex: number | null, 
        lastMemberId: number | null, 
        members: Omit<Member, 'isActive'>[] 
    } | null;
}
```

**Why chosen**:
- Separates rotation algorithm from service logic
- Easy to add new strategies (Random, Weighted, etc.) without changing existing code
- Improves testability by allowing mock strategies

### 2. **Singleton Pattern** 🔒
**File**: `src/services/RotationService.ts`

```typescript
export class RotationService {
    private static instance: RotationService;
    private lastSelectedMemberIndex: number | null = null;
    private lastSelectedMemberId: number | null = null;
    
    static getInstance(): RotationService {
        if (!RotationService.instance) {
            RotationService.instance = new RotationService();
        }
        return RotationService.instance;
    }
}
```

**Why chosen**:
- Ensures single source of truth for rotation state
- Maintains `lastSelectedMemberIndex` and `lastSelectedMemberId` consistently across all requests
- Prevents duplicate instances in the application

### 3. **Factory Pattern** 🏭
**File**: `src/strategies/RotationStrategy.ts`

```typescript
export class RotationStrategyFactory {
    private static roundRobinStrategyInstance: RotationStrategy;

    static getRoundRobinStrategy(): RotationStrategy {
        if (!this.roundRobinStrategyInstance) {
            this.roundRobinStrategyInstance = new RoundRobinStrategy();
        }
        return this.roundRobinStrategyInstance;
    }
}
```

**Why chosen**:
- Centralizes strategy creation
- Ensures single strategy instance (combines with Singleton)
- Makes it easy to add new strategy types

### 4. **Layered Architecture** 🎯

```
Request → Controller → Service → Strategy → Response
```

- **Controller**: Handles HTTP request/response and input validation
- **Service**: Contains business logic and manages state
- **Strategy**: Implements rotation algorithm
- **Data**: Centralized member storage

## 🧪 Testing Coverage

### Test Cases Implemented (10 tests)

| # | Test Case | Challenge Requirement |
|---|-----------|----------------------|
| 1 | Error when count is invalid number (< 1) | Input validation |
| 2 | Error when count is not a number | Input validation |
| 3 | Default count of 1 when not provided | Default behavior |
| 4 | Return multiple members when count > 1 | Example 4: Next N members |
| 5 | Error if count > member list length | Edge case validation |
| 6 | No immediate repetition | Example 2: No repetition |
| 7 | Error when not enough active members | Edge case |
| 8 | Error when no active members available | Example 6: All inactive |
| 9 | Correct behavior when index and ID are out of sync | Edge case recovery |
| 10 | Error if member ID not found in list | Error handling |

### Running Tests

```bash
yarn test
```

**Expected Output:**
```
 PASS  src/__tests__/component.test.ts
  RotationController
    ✓ should return error when count is not valid number
    ✓ should return error when count is not a number
    ✓ should use default count of 1 when not provided
    ✓ should return multiple members when count > 1
    ✓ should return error if count > length of memberList
    ✓ should have no immediate repetition
    ✓ should return error when not enough active members
    ✓ should return error when no active members available
    ✓ should return correct result when member index and ID are out of sync
    ✓ should throw error if member ID not found in list

Test Suites: 1 passed, 1 total
Tests:       10 passed, 10 total
```

## 📊 Design Decisions & Trade-offs

### Decision 1: Track Both Index and ID
**Choice**: Store both `lastSelectedMemberIndex` and `lastSelectedMemberId`

**Pros**:
- ✅ Fast O(1) lookup using index for next rotation
- ✅ ID provides data integrity check when member list changes
- ✅ Self-healing: recovers automatically if index/ID become out of sync

**Cons**:
- ❌ Slightly more complex than tracking id alone

### Decision 2: In-Memory State with Centralized Data File
**Choice**: Use `getMemberlist()` function from `src/data/members.ts`

**Pros**:
- ✅ Simplicity and fast access
- ✅ Meets challenge requirements (no database required)
**Cons**:
- ❌ Data lost on server restart
- ❌ Does not scale to distributed systems
- ❌ No persistence for audit trails

**Cons**

### Decision 3: Strategy + Factory Patterns
**Choice**: Use Strategy pattern with Factory for rotation logic

**Pros**:
- ✅ Clean separation of concerns
- ✅ Easy to extend with new algorithms

**Cons**:
- ❌ Added complexity with multiple abstraction layers

### Decision 4: Exclude `isActive` from Response
**Choice**: Return `{ id, name }` without `isActive` field

**Pros**:
- ✅ Cleaner API response (only relevant data)
- ✅ Internal state not exposed to clients

**Cons**:
- ❌ Client cannot see member status (but not needed for rotation result)

## 📋 Default Team Members

Defined in `src/data/members.ts`:

| ID | Name | Status |
|----|------|--------|
| 1 | Alice | ✅ Active |
| 2 | Bob | ✅ Active |
| 3 | Charlie | ✅ Active |
| 4 | Diana | ✅ Active |

## 🔄 How It Works

### Rotation Flow

1. **Request**: Client calls `GET /api/next?count=N`
2. **Controller**: Validates input (count >= 1) and extracts count parameter
3. **Service**: Gets member list via `getMemberlist()` and calls strategy
4. **Strategy**: 
   - Filters active members
   - Handles index/ID sync if needed
   - Selects next N members in round-robin order
5. **Track**: Service updates `lastSelectedMemberIndex` and `lastSelectedMemberId`
6. **Response**: Returns selected members `{ id, name }`

### Example Sequence

```
Call 1: GET /api/next       → Alice
Call 2: GET /api/next       → Bob
Call 3: GET /api/next       → Charlie
Call 4: GET /api/next       → Diana
Call 5: GET /api/next       → Alice (rotation restarts)
```

```
Call 1: GET /api/next?count=2  → [Alice, Bob]
Call 2: GET /api/next?count=2  → [Charlie, Diana]
Call 3: GET /api/next?count=2  → [Alice, Bob]
```

### Index-ID Sync Recovery

The strategy handles edge cases where `lastSelectedMemberIndex` and `lastSelectedMemberId` become out of sync:

```typescript
// If index points to Bob (id:2) but lastMemberId is Charlie (id:3)
// Strategy finds Charlie's actual index and continues from there
if (members[lastMemberIndex].id !== lastMemberId) {
    const correctLastIndex = members.findIndex(m => m.id === lastMemberId);
    currentIndex = (correctLastIndex + 1) % numberLength;
}
```

## 🚀 Future Enhancements (With More Time)

1. **Persistent Storage**: Add database layer (PostgreSQL/MongoDB)
2. **Full History**: Track complete rotation history for analytics
3. **Multiple Strategies**: Add Random, Weighted, and Load-Based rotation
4. **API Validation**: Add Joi or Zod for request validation
5. **Swagger Documentation**: Auto-generated API docs
6. **Metrics**: Track rotation fairness metrics over time
7. **Cache**: Use Redis as cache for lastSelectedMemberIndex and lastSelectedMemberId to support the distributed system

**Built with ❤️ for the TWVN Coding Challenge 2026**