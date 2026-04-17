<<<<<<< HEAD
# Coding Challenge 2026

Welcome to the TWVN Coding Challenge! Whether you're a seasoned developer or just starting your journey, this event is designed for **YOU**. This is a safe, fun space to experiment with new technologies, learn from your peers, and grow your skills — all while having a great time!

> **New to coding challenges?** Don't worry! We welcome all skill levels. Partial solutions, learning attempts, and "I tried something new" are all celebrated here.
=======
# SMS Service

A minimal, well-tested HTTP service for routing and tracking outbound SMS messages across multiple providers and carriers, built for a cinema booking ecosystem.
>>>>>>> 3f04f81 (add models and repositories for SMS service components)

---

## 📋 Table of Contents

<<<<<<< HEAD
1. [🌟 Why Join?](#why-join)
2. [🚀 Quick Start: How to Join](#quick-start-how-to-join)
3. [📏 Rules & Guidelines](#rules--guidelines)
4. [📤 How to Submit Your Code](#how-to-submit-your-code)
5. [🏆 Scoring, Leaderboard & Prizes](#scoring-leaderboard--prizes)
6. [🎤 Demo Day](#demo-day)
7. [💡 Tips for First-Timers](#tips-for-first-timers)
8. [❓ FAQ](#faq)
9. [💬 Communication & Contact](#communication--contact)

---

## 🌟 Why Join?

- 🚀 **Learn by doing** — Try that framework or pattern you've been curious about
- 🤝 **Connect with peers** — Work with people across teams you might not meet otherwise
- 🏆 **Build confidence** — No pressure, just growth and friendly competition
- 🎉 **Have fun** — Seriously, this is meant to be enjoyable!

### What You'll Get

- Knowledge Badges for your profile (e.g., 🟢 _TDD Specialist_, 🟢 _Clean Architecture_)
- A chance to present to the dev community on Demo Day
- **Team vouchers** for 1st, 2nd, and 3rd place at **each** coding challenge
- **Season Champion** — highest Season Score wins a **Mac mini M4** at the end of the season

---

## 🚀 Quick Start: How to Join

| Step                  | What to Do                                                                                                                                |
| --------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| **1. Register**       | Drop a message in `#Vietnam Dev Community` ([join here](https://chat.google.com/room/AAAAm0xrYkU?cls=7))                                  |
| **2. Form a Team**    | Join solo, form a team of 2-3 people, or ask in the channel for teammates                                                                 |
| **3. Attend Kickoff** | Get the challenge brief, repo access, and timeline                                                                                        |
| **4. Code & Learn**   | Spend 2-3 hours/week coding. Ask questions in the channel anytime                                                                         |
| **5. Submit**         | Fork the repo, implement in your fork, then open a Pull Request back to the official repo (see [How to Submit](#how-to-submit-your-code)) |
| **6. Present**        | Demo your solution on Demo Day and celebrate!                                                                                             |

---

## 📏 Rules & Guidelines

### 👥 Teams

- **Solo or teams of 1-3** — your choice
- **Pairing encouraged:** mix senior + junior intentionally
- Rotate pairs every challenge for variety

### ⏱️ Time Commitment

- **2-3 hours per week** over a 2-3 week coding period
- Partial solutions are welcome — "good enough" beats "perfect"

### ✅ What's Allowed

| ✅ Allowed                                                    | ❌ Not Allowed                  |
| ------------------------------------------------------------- | ------------------------------- |
| Any language/framework (unless specified)                     | Client data or proprietary code |
| AI tools (Copilot, ChatGPT, Cursor) — **bonus if documented** | Over-engineering beyond scope   |
| Open-source libraries                                         |                                 |

---

## 📤 How to Submit Your Code

You submit using a **fork-based workflow**:

1. The organizers create **one official challenge repository**.
2. Teams **fork** it and implement the solution in their own fork.
3. Teams submit by opening a **Pull Request back to the official repo** (or share the fork link as a fallback).

Follow the steps below in order.

### 🍴 Step 1: Fork the Official Repository

- Fork the official repo to your own GitHub account/org.
- All your work happens in your fork (not in the official repo).

### 💻 Step 2: Clone Your Fork

- Clone your fork locally.
- Add the upstream remote (optional but recommended) so you can pull updates.

### 🔀 Step 3: Create Your Working Branch (Recommended)

Avoid committing directly to your fork's `main`. Create a working branch right after you clone.

**Branch naming format (keep this):**

```
coding-challenge-{number}/<your-team-name>
```

**Examples:** `coding-challenge-1/rustaceans`, `coding-challenge-1/stevelam`

### 📁 Step 4: Add Your Code to Your Team Folder

Place **all** your code under your team's dedicated folder:

```
coding-challenge-{number}/<your-team-name>/
```

**Examples:**

- `coding-challenge-1/rustaceans/` — Team folder for the Rustaceans team
- `coding-challenge-1/stevelam/` — Individual folder for Steve Lam

**Your folder should contain:**

- All source code
- `README.md` (use the template below)
- Tests
- Any configuration files

#### README Template (for your team folder)

```markdown
# [Your Team Name] - [Challenge Name]

## Team Members

- Developer 1 (role/team)
- Developer 2 (role/team)

## How to Run

[Step-by-step instructions]

## Approach

[Brief explanation of your solution]

## Design Decisions

- Why did you choose this tech stack?
- What trade-offs did you make?
- What patterns did you use?

## Challenges Faced

[What was hard? How did you overcome it?]

## What We Learned

[New skills, technologies, or insights]

## With More Time, We Would...

[Nice-to-haves you didn't implement]

## AI Tools Used (if any)

[Which tools? How did they help?]
```

### 🔃 Step 5: Open a Pull Request to the Official Repo (Required) ⚠️

**Your PR is your official submission.** This is mandatory when possible.

1. Push your branch to **your fork**
2. Open a Pull Request from `your-fork:your-branch` into `official-repo:main`
3. Fill out the PR description using the template below
4. Request review from the challenge organizers or mentors
5. **Do not merge** — Organizers will review all PRs before Demo Day

**Fallback (only if PR is not possible):** Share the link to your fork (and the branch) with the organizers.

**PR Title Format:**

```
[Challenge #X] Team Name - Brief Solution Description
```

**PR Description Template:**

```markdown
## Team Information

- **Team Name**: [Your Team Name]
- **Members**:
  - [Name 1] - [Role/Team]
  - [Name 2] - [Role/Team]

## Solution Overview

[Brief description of your approach - 2-3 sentences]

## Key Design Decisions

- [Decision 1 and why]
- [Decision 2 and why]
- [Design patterns used]

## Technologies & Tools

- Language/Framework: [e.g., Python, Node.js, Rust]
- Key Libraries: [List main dependencies]
- AI Tools Used: [e.g., GitHub Copilot, ChatGPT - if any]

## How to Run

[Step-by-step instructions to run your code]

## How to Test

[Instructions to run tests if available]

## Trade-offs & Limitations

[What you optimized for and what was sacrificed]

## With More Time, We Would...

[Features or improvements you'd add]
=======
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

<<<<<<< HEAD
## 🏆 Scoring, Leaderboard & Prizes

> **TL;DR** — Each challenge is scored out of 100 (+ bonus). Points are tracked per individual across each season of **four** coding challenges; your **Season Score** uses your **top three** challenge scores only. The highest Season Score wins the **Special Prize** (Mac mini M4). **Teams** placing 1st, 2nd, and 3rd at each challenge still receive **vouchers** for that round.

### How Points Work for Teams vs Individuals

| Model            | How It Works                                                                                               |
| ---------------- | ---------------------------------------------------------------------------------------------------------- |
| **Individual**   | You code alone. Your score goes directly to your profile.                                                  |
| **Team (max 3)** | The team is graded as one unit. Every member receives the full team score. Bonus points apply to everyone. |

All points are tracked **per individual** for the leaderboard, regardless of whether you join solo or in a team.

---

### Per-Challenge Scoring (Max 100 + Bonus)

#### A. Core Execution & Design — 70 Points

This evaluates the functional correctness and architectural quality of your submission. Each challenge defines its own detailed breakdown within this block (see the specific challenge document).

#### B. Tests & Reliability — 15 Points

| Criteria                                                    | Points |
| ----------------------------------------------------------- | ------ |
| **Coverage** — Tests covering the core logic                | 10     |
| **Resilience** — Tests for edge cases, errors, empty states | 5      |

#### C. Documentation & Communication — 15 Points

| Criteria                                                         | Points |
| ---------------------------------------------------------------- | ------ |
| **README** — Clear setup instructions and trade-off explanations | 5      |
| **Presentation** — Ability to explain _why_ decisions were made  | 10     |

---

### Bonus Points (Up to +25)

These let you exceed 100 and are the fast track to winning the Special Prize.

| Bonus                | Points    | How to Earn                                                                      |
| -------------------- | --------- | -------------------------------------------------------------------------------- |
| **TDD Approach**     | +5        | Commit history proves tests written _before_ implementation (Red-Green-Refactor) |
| **Junior-Led Demo**  | +5        | The most junior team member presents the solution                                |
| **Community Assist** | +5        | Provide constructive peer reviews on other participants' PRs                     |
| **Creative Bonus**   | up to +10 | Subjective — decided by the facilitator or community vote                        |

### Community Engagement Points (Demo Day)

Earned during Demo Day and added to your **Season Total**.

| Award                         | Points      | Details                                                       |
| ----------------------------- | ----------- | ------------------------------------------------------------- |
| **"Most Loved" Solution**     | +10         | Voted by the audience via live poll                           |
| **The Curious Mind**          | +2/question | Ask a meaningful technical question during Q&A (max +6/month) |
| **The Subject Matter Expert** | +2          | Give an exceptionally clear answer during your own Q&A        |

---

### The Long-Run Leaderboard (Seasons)

Your **Season Score** determines who wins the **Special Prize** for the season: a **Mac mini M4**.

Each **season** runs for **four** coding challenges. Your season score is the sum of your **three highest** challenge scores from that season—the fourth (lowest) score is dropped. If you complete **fewer than four** challenges, there is no extra score to drop: your season score is still the sum of your **three highest** scores among the challenges you entered (or all of your scores if you have fewer than three). There are no separate participation or placement bonuses on top of this.

**The "Best of 3 (of 4)" Rule:**

```
Season Score = (Top Score 1) + (Top Score 2) + (Top Score 3)
```

<details>
<summary><strong>Example: Alice vs Bob (click to expand)</strong></summary>

**Alice (The Consistent Performer)** — Completes all four challenges; one rough round is dropped.

| Challenge | Score  |
| --------- | ------ |
| C1        | 80     |
| C2        | **85** |
| C3        | **82** |
| C4        | 45     |

- Top 3 Scores: 85 + 82 + 80 = **247** (45 dropped)
- **Season Total: 247**

---

**Bob (Three Strong Rounds)** — Only takes part in three challenges this season—every score counts; there is no fourth score to drop.

| Challenge | Score  |
| --------- | ------ |
| C1        | **95** |
| C2        | **90** |
| C3        | **88** |
| C4        | —      |

- Top 3 Scores: 95 + 90 + 88 = **273**
- **Season Total: 273**

---

**Comparison**

|                  | Alice (4 challenges) | Bob (3 challenges) |
| ---------------- | -------------------- | ------------------ |
| **Top 3 Scores** | 247                  | 273                |
| **SEASON SCORE** | **247**              | **273**            |

> Takeaway: Having four scores lets you drop your weakest one. With only three scores, all of them count toward the same top-three cap—so one off week can’t be discarded.

</details>

---

### 🎁 Prizes & Recognition

**Each coding challenge:**

- 🎟️ **Team vouchers (1st, 2nd, 3rd)** — The top three **teams** for that challenge each receive vouchers according to placement (1st, 2nd, 3rd). This is unchanged: even with the season-long leaderboard, every challenge still has its own team podium and prizes.
- 🏅 Knowledge Badges for skills demonstrated (e.g., 🟢 _TDD Specialist_, 🟢 _Clean Architecture_)
- 🎤 Top 3 get the Demo Slot to present to the entire dev community

**End of each season:**

- 🏆 **The Season Champion (Special Prize)** — Highest **Season Score** wins a **Mac mini M4**

---

### 📅 Monthly Workflow

| Week         | Activity                                             |
| ------------ | ---------------------------------------------------- |
| **Week 1**   | Challenge Release & Registration                     |
| **Week 2-3** | Hacking & Development (PRs due by end of Week 3)     |
| **Week 4**   | Facilitator Review & Scoring                         |
| **Week 5**   | Demo Day & Winner Announcement (new challenge drops) |

---

## 🎤 Demo Day

Each team gets **10-15 minutes** to present:

| Time  | Activity                      |
| ----- | ----------------------------- |
| 1 min | Problem recap                 |
| 5 min | Live demo or code walkthrough |
| 3 min | Design decisions & trade-offs |
| 2 min | Lessons learned               |

You can also prepare:

- A 5-minute live demo OR recorded video
- Slides (optional, not required)

---

## 💡 Tips for First-Timers

### 🎯 Before You Start

1. **Read the challenge carefully** — Understand the requirements before coding
2. **Start simple** — Get something working, then improve it
3. **Don't aim for perfect** — "Good enough" is the goal
4. **Ask questions early** — Don't struggle alone for days

### 💻 During Development

1. **Time-box your work** — Set a timer for 2-3 hour blocks
2. **Commit frequently** — Small commits are easier to manage
3. **Document as you go** — Write down your decisions and trade-offs
4. **Test your solution** — Even basic tests count!

### 🎬 For Demo Day

1. **Practice your demo** — Run through it once before presenting
2. **Focus on learning** — Share what you learned, not just what you built
3. **Be honest** — "I tried X but it didn't work" is valuable!
4. **Have fun** — This isn't a job interview, it's a celebration

---

## ❓ FAQ

**Q: I'm a junior developer. Is this too advanced for me?**
A: Not at all! This event is designed for **all levels**. You'll be paired with others, and learning is valued as much as the solution itself. Many juniors have won past challenges!

**Q: What if I can't finish in time?**
A: Partial solutions are 100% welcome! Submit what you have and explain what you would do next.

**Q: What if my solution breaks during the demo?**
A: It happens! Explain what was supposed to happen. Judges value your understanding and design decisions over a perfect demo.

**Q: How much time should I really spend?**
A: Aim for **4-6 hours total** over 2-3 weeks. That's roughly 2 hours per week. Don't overwork — this should be fun!

**Q: Can I use code from previous projects?**
A: You can reuse your own patterns and approaches, but the solution should be built for this specific challenge. No copy-pasting entire projects.

**Q: What languages/frameworks are allowed?**
A: Unless the challenge specifies, use whatever you want! Want to try Rust? Go for it. Comfortable with Python? Perfect!

**Q: Do I need to use AI tools?**
A: Nope! It's optional. If you do use them (GitHub Copilot, ChatGPT, Cursor, etc.), document how they helped for bonus points.

**Q: What if I have a question not listed here?**
A: Ask in `#Vietnam Dev Community` ([join here](https://chat.google.com/room/AAAAm0xrYkU?cls=7)) or message the organizers. No question is too small!

---

## 💬 Communication & Contact

**📢 Main Channel:** `#Vietnam Dev Community` ([join here](https://chat.google.com/room/AAAAm0xrYkU?cls=7))
Use this for questions, teammate search, and general discussion.

**🆘 Need Help?**

- Senior devs volunteer as mentors — just ask!
- Stuck? Post in the channel, we're all here to help.

**📞 Organizers:**

- Steve Lam — <tai.lam@thoughtworks.com>
- Tom Tang — <tom.tang@thoughtworks.com>
- Technical Mentors — Check `#Vietnam Dev Community` pinned messages

**Have ideas to improve the event?** DM the organizers or share in the retro session.

---

## 🚀 Ready to Start?

1. ✅ Join the `#Vietnam Dev Community` space ([join here](https://chat.google.com/room/AAAAm0xrYkU?cls=7))
2. ✅ Mark the kickoff event in your calendar
3. ✅ Start thinking about who you'd like to team up with
4. ✅ Get excited — this is going to be fun! 🎉

**See you at the challenge!** 🚀
=======
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
>>>>>>> 3f04f81 (add models and repositories for SMS service components)
