
# Coding challenge 2
- [Overview](#-overview)
- [Functional Overview](#functional-overview)
- [Project Structure](#-project-structure)
- [Tech Stack](#-tech-stack)
- [Installation](#-installation)
- [Usage](#-usage)
- [API Design](#api-design)
- [Business Logic](#-business-logic)
- [Aggregate Design](#aggregate-design)
- [Design Patterns](#-design-patterns)
- [Testing](#-testing)
- [API Examples](#-api-examples)
- [Out of Scope](#out-of-scope)
- [Assumptions](#-assumptions)
- [Trade-offs](#-trade-offs)

---

## Functional Overview

The SMS Service is a reusable, shared platform capability that enables multiple domains in the cinema booking ecosystem (Customer, Booking, ERP, Accounting/Operations, Data) to send SMS messages and track delivery and cost outcomes in a consistent way across multiple countries and SMS providers.

The SMS Service is responsible for:

- Accepting SMS send requests (e.g., OTP during membership registration, booking confirmations, operational notifications).
- Determining the recipient's carrier (MNO) from the phone number (simulation is allowed).
- Selecting exactly one SMS provider using current routing rules based on country + carrier. We perform provider selection; downstream carrier routing is out of scope / handled by the provider.
- Managing the SMS message lifecycle states and transitions in a controlled, traceable way.
- Capturing estimated cost at send-to-provider time and actual cost once delivery succeeds.
- Handling asynchronous provider callbacks (simulated) to update message state and actual cost.
- Providing basic observability via logs and in-memory reporting to support cost and outcome analysis.

---

## 📁 Project Structure

```
twvn-coding-challenge-2/
├── cmd/
│   └── api/
│       └── main.go                         # Entry point: wires dependencies and starts the server
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   │   ├── dto.go                      # Request/response DTO types
│   │   │   ├── sms_message_handler.go      # HTTP handlers: SendSMS, Callback, GetByID
│   │   │   └── sms_message_handler_test.go # Unit tests for handlers
│   │   └── routes/
│   │       └── routes.go                   # Route registration (Gin)
│   ├── models/                             # Domain models and error types
│   │   ├── sms_message.go                  # SMSMessage aggregate + status transitions
│   │   ├── carrier.go
│   │   ├── country.go
│   │   ├── provider.go
│   │   ├── provider_agreement.go
│   │   ├── provider_selection_decision.go
│   │   ├── sender.go
│   │   ├── money.go
│   │   └── errors.go
│   ├── repositories/                       # In-memory repository implementations
│   │   ├── sms_message_repository.go
│   │   ├── carrier_repository.go
│   │   ├── country_repository.go
│   │   ├── provider_repository.go
│   │   ├── provider_agreement_repository.go
│   │   ├── sender_repository.go
│   │   ├── provider_selection_decision_repository.go
│   │   └── *_test.go                       # Unit tests per repository
│   └── services/
│       ├── provider_selector.go            # GetFirstProviderSelector implementation
│       ├── provider_selector_test.go       # Unit tests for provider selector
│       └── api_clients/
│           ├── provider_client.go          # ProviderAPIClient interface + factory
│           ├── twilio_client.go            # Twilio stub (fixed cost: 500 VND)
│           ├── vonage_client.go            # Vonage stub (fixed cost: 1000 VND)
│           └── provider_client_test.go     # Unit tests for API clients
├── tests/
│   └── integration/
│       ├── send_sms_test.go                # Integration test: full send flow
│       └── provider_callback_update_test.go # Integration tests: callback flows
├── go.mod
├── go.sum
└── README.md
>>>>>>> 3f04f81 (add models and repositories for SMS service components)
```

---


## 🛠️ Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.26 |
| HTTP Framework | Gin v1.12 |
| Testing | Testify v1.11 (assert, require) |
| Database | In-memory (thread-safe, `sync.RWMutex`) |
| Dependency Management | Go Modules |

**Third-party libraries:**

- `github.com/gin-gonic/gin` — HTTP router and middleware
- `github.com/stretchr/testify` — Assertions and test utilities
- `gorm.io/gorm` — ORM (imported, ready to wire a real DB)
- `github.com/joho/godotenv` — Environment variable loading

---

## 🚀 Installation

### Prerequisites

- **Go** >= 1.21 ([download](https://go.dev/dl/))

### Setup

**Clone the repository:**

```bash
git clone <repo-url>
cd twvn-coding-challenge-2
```

**Install dependencies:**

```bash
go mod download
```

**Run the server:**

```bash
go run cmd/api/main.go
```

The server starts on `http://localhost:8080`.

**Run all tests:**

```bash
go test ./...
```

**Run tests with coverage:**

```bash
go test ./... -cover -vet=off
```

---

## 💻 Usage

### Send an SMS

```bash
curl -s -X POST http://localhost:8080/api/v1/sms/send \
  -H 'Content-Type: application/json' \
  -d '{
    "sender_id": "sender-001",
    "recipient_phone": "931234567",
    "content": "Your OTP is 123456",
    "country_code": "84"
  }'
```

**Response:**

```json
{
  "data": {
    "message_id": "msg-1713340800000000000",
    "status": "SEND_TO_PROVIDER",
    "provider_id": "provider-twilio",
    "carrier_id": "carrier-001",
    "estimated_cost": "500",
    "currency": "VND",
    "created_at": "2026-04-17T07:12:00Z"
  },
  "meta": { "request_id": "req-...", "timestamp": "..." }
}
```

### Get a message by ID

```bash
curl -s http://localhost:8080/api/v1/sms/<MESSAGE_ID>
```

### Simulate a provider webhook callback

```bash
curl -s -X POST http://localhost:8080/api/v1/sms/webhooks/provider-callback \
  -H 'Content-Type: application/json' \
  -d '{
    "provider_id": "provider-twilio",
    "message_id": "<MESSAGE_ID>",
    "status": "QUEUE",
    "event": "queued",
    "occurred_at": "2026-04-17T07:12:00Z"
  }'
```

---

## API Design

- `POST /api/v1/sms/send` — Send an SMS message
- `GET /api/v1/sms/:id` — Retrieve an SMS message by ID
- `POST /api/v1/sms/webhooks/provider-callback` — Receive a delivery status update from a provider

---

## 🧠 Business Logic

### SMS Message Lifecycle

Every SMS message moves through a strict, enforced state machine:

```
NEW
 └─► SEND_TO_PROVIDER
       └─► QUEUE
             ├─► SEND_TO_CARRIER
             │     ├─► SEND_SUCCESS   (terminal — actual cost recorded)
             │     └─► SEND_FAILED    ──► SEND_TO_PROVIDER  (client has to resend the message)
             └─► CARRIER_REJECTED     ──► SEND_TO_PROVIDER  (client has to resend the message)
```

Transitions are validated by `SMSMessage.CanTransitionTo()`. Any invalid transition is rejected with a `400 INVALID_STATUS` response.

### Provider Selection

1. Resolve the recipient's **country** from the `country_code` field.
2. Detect the recipient's **carrier** by matching the phone number prefix against known carrier prefixes for that country.
3. Look up all **provider agreements** for that carrier.
4. Select a provider from the list based on business rules.
5. Call the provider's client to get an **estimated cost**.
6. Record a `ProviderSelectionDecision` and update the message status to `SEND_TO_PROVIDER`.

### Cost Tracking

| Event | Action |
|---|---|
| Provider selected | `estimated_cost` written to message |
| Callback: `SEND_SUCCESS` | `actual_cost` written from callback payload |
| Callback: `SEND_FAILED` / `CARRIER_REJECTED` | `failure_reason` written; actual cost left null |

---

## Aggregate Design

The system uses an in-memory database but it can be easily replaced with a real database like Postgres by using the repository pattern.

**SMS Message**
- Id
- SenderId
- RecipientId
- Content
- Status (`NEW`, `SEND_TO_PROVIDER`, `QUEUE`, `SEND_TO_CARRIER`, `SEND_SUCCESS`, `SEND_FAILED`, `CARRIER_REJECTED`)
- Estimated Cost
- Actual Cost

**Sender**
- Id
- Phone number

**Recipient**
- Id
- Phone number

**Country**
- Id
- Phone number prefix

**Provider**
- Id
- Name

**Carrier**
- Id
- Name
- CountryId
- Phone number prefix

**Provider Agreement**
- Id
- CarrierId
- ProviderId

**Provider Selection Decision**
- SMS Message Id
- Estimated Cost
- ProviderId

---

## 🎨 Design Patterns

### Repository Pattern

Every data entity has its own `interface` (e.g. `SMSMessageRepository`, `CarrierRepository`) with an in-memory implementation. The handlers and services depend only on the interface, making it trivial to replace the in-memory store with a Postgres/GORM implementation without touching any business logic.

### Strategy Pattern — Provider Selector

The `ProviderSelector` interface decouples selection logic from the rest of the system:

```go
type ProviderSelector interface {
    Select(message *models.SMSMessage) (models.ProviderSelectionDecision, error)
}
```

`GetFirstProviderSelector` is the current implementation. New strategies (e.g. cheapest-first, round-robin, load-balanced) can be added by implementing this interface without changing any handler code.

### Dependency Injection

All dependencies (`SMSMessageRepo`, `CountryRepo`, `CarrierRepo`, etc.) are injected into `SMSMessageHandler` through `SMSMessageHandlerDeps`. This is the primary mechanism that makes unit testing straightforward — tests pass real in-memory repositories or swap them for controlled instances.

### Composition over Inheritance

The handler struct holds a `deps` value rather than embedding or extending types. Service components are composed together in `main.go` explicitly, keeping the dependency graph visible and flat.

---

## 🧪 Testing

This project follows a **Test-Driven Development (TDD)** approach. Tests are written alongside production code to drive design decisions, catch regressions early, and document expected behaviour.

### Test Coverage

| Package | Coverage |
|---|---|
| `internal/repositories` | **100%** |
| `internal/services/api_clients` | **100%** |
| `internal/services` | **94.7%** |
| `internal/api/handlers` | **86.1%** |

### Running tests

```bash
# All tests
go test ./...

# With verbose output
go test ./... -v

# With coverage summary
go test ./... -cover -vet=off

# With HTML coverage report
go test ./... -coverprofile=coverage.out -vet=off
go tool cover -html=coverage.out

# Per package
go test ./internal/repositories/... -v
go test ./internal/services/... -v -vet=off
go test ./internal/api/handlers/... -v
go test ./tests/integration/... -v
```

---

## 📚 API Examples

### Health check

```bash
curl -s http://localhost:8080/health
```

### Full message lifecycle (run in sequence)

**1 — Send SMS:**

```bash
MSG_ID=$(curl -s -X POST http://localhost:8080/api/v1/sms/send \
  -H 'Content-Type: application/json' \
  -d '{"sender_id":"sender-001","recipient_phone":"931234567","content":"Your OTP is 123456","country_code":"84"}' \
  | jq -r '.data.message_id')
echo "Message ID: $MSG_ID"
```

**2 — Provider queues the message:**

```bash
curl -s -X POST http://localhost:8080/api/v1/sms/webhooks/provider-callback \
  -H 'Content-Type: application/json' \
  -d '{"provider_id":"provider-twilio","message_id":"'"$MSG_ID"'","status":"QUEUE","event":"queued","occurred_at":"2026-04-17T07:12:00Z"}'
```

**3 — Provider dispatches to carrier:**

```bash
curl -s -X POST http://localhost:8080/api/v1/sms/webhooks/provider-callback \
  -H 'Content-Type: application/json' \
  -d '{"provider_id":"provider-twilio","message_id":"'"$MSG_ID"'","status":"SEND_TO_CARRIER","event":"sent_to_carrier","occurred_at":"2026-04-17T07:12:01Z"}'
```

**4a — Delivery success (records actual cost):**

```bash
curl -s -X POST http://localhost:8080/api/v1/sms/webhooks/provider-callback \
  -H 'Content-Type: application/json' \
  -d '{"provider_id":"provider-twilio","message_id":"'"$MSG_ID"'","status":"SEND_SUCCESS","event":"delivery_success","occurred_at":"2026-04-17T07:12:02Z","actual_cost":"450","currency":"VND"}'
```

**4b — Delivery failure:**

```bash
curl -s -X POST http://localhost:8080/api/v1/sms/webhooks/provider-callback \
  -H 'Content-Type: application/json' \
  -d '{"provider_id":"provider-twilio","message_id":"'"$MSG_ID"'","status":"SEND_FAILED","event":"delivery_failed","occurred_at":"2026-04-17T07:12:02Z","failure_reason":"Number unreachable"}'
```

**4c — Carrier rejected:**

```bash
curl -s -X POST http://localhost:8080/api/v1/sms/webhooks/provider-callback \
  -H 'Content-Type: application/json' \
  -d '{"provider_id":"provider-twilio","message_id":"'"$MSG_ID"'","status":"CARRIER_REJECTED","event":"carrier_rejected","occurred_at":"2026-04-17T07:12:02Z","failure_reason":"Carrier rejected the message"}'
```

**5 — Inspect message state:**

```bash
curl -s http://localhost:8080/api/v1/sms/$MSG_ID | jq
```

---

## Out of Scope

- **External integrations:** Real SMS provider SDK integration, Carrier lookup API, Real database
- **Data persistence:** State is lost on process restart
- **Infrastructure:** Real message queue systems, containerization, cloud deployment, background workers
- **Reliability engineering:** Retry infrastructure, circuit breaker, fault tolerance at infrastructure level
- **Performance tuning:** Caching, load balancing, scaling
- **Security**
- **Observability:** Real monitoring systems, metrics, distributed tracing, structured logging

---

## 📌 Assumptions

The following assumptions were made to keep the implementation focused and within scope:

1. **Currency is fixed as VND.** All cost values (estimated and actual) use VND. No multi-currency conversion logic is implemented.

2. **All carrier-provider pairs are supported.** Any carrier can be routed to any provider as long as a `ProviderAgreement` record exists. There is no restriction based on geography or contract type.

3. **Provider callback statuses match the defined status constants.** The webhook `status` field in callback payloads is expected to be one of the codebase's `SMSStatus` constants (`QUEUE`, `SEND_TO_CARRIER`, `SEND_SUCCESS`, `SEND_FAILED`, `CARRIER_REJECTED`).

4. **All provider callback APIs return the same payload format.** A single `ProviderCallbackRequest` DTO handles callbacks from all providers (Twilio, Vonage, etc.) without provider-specific parsing.

5. **All external service implementations are out of scope.** Provider clients (`TwilioAPIClient`, `VonageAPIClient`) are stubs that return hardcoded responses. No real HTTP calls are made to provider APIs.

6. **Provider selector algorithm is intentionally minimal.** The `GetFirstProviderSelector` simply picks the first agreement found. The `ProviderSelector` interface is designed for extensibility — more sophisticated strategies (cheapest-first, round-robin, load-balanced) can be added without touching existing code.

7. **The real-world message flow is followed end-to-end.** The service models the actual lifecycle: a sender initiates a message → a provider is chosen based on business routing rules → the provider processes and relays to the carrier → the carrier delivers → the service is notified via callback. Each step is a traceable status transition.

8. **In-memory storage replaces a real database.** All repositories are thread-safe in-memory maps and slices. They implement the same interfaces that a GORM/Postgres implementation would, so switching storage backends requires only a new implementation of the relevant repository interface — no handler or service changes needed.

---

## ⚖️ Trade-offs

### Procedural Style vs OOP

**Decision:** Business logic in the handlers and services is written in a procedural, step-by-step style rather than using deep class hierarchies or domain-rich objects.

**Rationale:** Go's idiomatic style favours explicit, readable functions over inheritance and polymorphism. The codebase is straightforward to trace: each handler method is a self-contained sequence of steps.

**Trade-off:** Some duplication in error-response construction and meta-building across handlers. An object or helper could reduce this, but the current approach keeps each handler easy to read in isolation without jumping between abstractions.

---

### Composition over Inheritance

**Decision:** Dependencies are composed through struct fields (`SMSMessageHandlerDeps`) and constructor injection (`NewSMSMessageHandler`, `NewGetFirstProviderSelector`), not through embedded types or base structs.

**Rationale:** Composition gives precise control over what each component depends on. Tests can substitute any single dependency without having to satisfy a deep inheritance chain. Adding a new dependency is a local change (add a field, update the constructor) rather than a refactor of a base type.

**Trade-off:** Constructors become more verbose as the dependency list grows. In practice, `main.go` is the only place where all dependencies are wired together, so the verbosity is contained and visible in one place.

---

### In-Memory Store vs Real Database

**Decision:** All repositories are backed by `sync.RWMutex`-protected Go maps and slices rather than a real database.

**Rationale:** A real database is explicitly out of scope. The repository pattern ensures that the rest of the codebase is completely unaware of the storage mechanism — the interfaces would be satisfied equally by a GORM implementation.

**Trade-off:** State is lost on process restart. Concurrent write load is limited by single-process Go mutex contention. Neither matters within the scope of this challenge.

---

### Manual Dependency Injection vs IoC Container

**Decision:** All dependencies are wired by hand in `main.go` using plain constructor calls, with no IoC framework (e.g. `wire`, `dig`, `fx`).

**Rationale:** For a service of this size, a DI framework adds more ceremony than it removes. The full wiring is visible in one place (`main.go`) and is easy to follow without knowing any framework-specific conventions. Go's explicit style makes manual wiring readable and unsurprising.

**Trade-off:** As the number of components grows, `main.go` becomes longer and changes to constructor signatures must be updated manually. An IoC container would generate or validate the wiring automatically. That cost is not justified at this scale.

---

### Single Provider Selection Strategy

**Decision:** Only one `ProviderSelector` implementation exists (`GetFirstProviderSelector`).

**Rationale:** Requirements specify a working, traceable routing flow, not a full routing engine. The `ProviderSelector` interface is the extension point — adding a cheapest-first or latency-based strategy is a new struct, not a change to existing code.

**Trade-off:** The current selection is deterministic but not cost-optimal. It always picks the provider with the lowest agreement index for a given carrier. This is acceptable for a simulation context.

