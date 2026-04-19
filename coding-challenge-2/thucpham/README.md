# SMS Gateway (Coding Challenge 2)

## Quick start

**Prerequisites:** [Node.js](https://nodejs.org/) (LTS recommended).

Run the interactive demo (terminal UI with `@clack/prompts`):

```bash
cd coding-challenge-2/thucpham
npx tsx main.ts
```

You will get a menu to send SMS, simulate provider callbacks, list messages, and view cost reports.

```
➜  users git:(coding-challenge-2/thucpham) ✗ npx ts-node main.ts
┌    SMS Gateway CLI
│
◆  What do you want to do?
│  ● Send SMS
│  ○ Simulate provider callback
│  ○ View all SMS
│  ○ Cost report
│  ○ Exit
└
```

**Tip:** send a message first, then use **Simulate provider callback** and step through the statuses to see the lifecycle and final result.

## Repo structure

Role of each folder / file (matches the current tree):

```
thucpham/
├── domain/                 ← Pure domain model, no I/O or framework coupling
│   ├── SmsMessage.ts       ← Entity: one message + audit + retry helpers
│   ├── SmsStatus.ts        ← Lifecycle status enum
│   ├── SmsAuditRecord.ts   ← One state-change audit row
│   └── StateMachine.ts     ← Allowed state transitions
│
├── routing/                ← Carrier detection + provider routing table
│   ├── CarrierDetector.ts  ← Infer carrier from phone number prefix
│   ├── RoutingTable.ts     ← Map (country, carrier) → ordered provider list
│   └── RoutingEngine.ts    ← Combines detector + table; resolves provider(s)
│
├── providers/              ← Send contract + simulated adapters
│   ├── SmsProvider.ts      ← Adapter interface (name, cost, send)
│   ├── Providers.ts        ← Concrete adapters (Twilio, Vonage, …)
│   └── ProviderRegistry.ts ← Register / resolve adapter by name
│
├── infrastructure/         ← Persistence, no heavy domain rules
│   └── SmsRepository.ts    ← In-memory Map id → SmsMessage + simple queries
│
├── application/            ← Use-case orchestration: send, callback, report, retry
│   ├── SmsService.ts       ← New send + retry after failure
│   ├── CallbackHandler.ts  ← Apply callback (new state) → state + optional retry
│   ├── RetryStrategy.ts    ← Same-provider / fallback strategies + resolver
│   └── CostReport.ts       ← Cost / volume / success rate from repository
│
├── cli.ts                  ← CLI menu — interactive “frontend” for the demo
├── main.ts                 ← Bootstrap: wire registry, routing, services, CLI
├── test/
│   └── flows.test.ts       ← Basic integration-style flow tests
├── vitest.config.ts
├── package.json
└── README.md
```

---

## Automated tests

From the `thucpham` directory (same folder as `package.json`):

```bash
npm test
```

Vitest runs `test/flows.test.ts` with the same wiring as `main.ts` (`buildSystem`). Covered flows:

- **Invalid country** — `send` throws (no routing / carrier data, e.g. `XX`).
- **Unknown carrier** — `send` throws when the number does not match a carrier for that country.
- **Happy path** — Viettel VN → Twilio; callbacks through queue/carrier to success; `actualCost` saved.
- **Same-provider retry** — `SEND_FAILED` + `RATE_LIMITED` → back on `SEND_TO_PROVIDER` with Twilio.
- **Fallback retry** — `SEND_FAILED` + `SERVICE_UNAVAILABLE` → retry on Vonage (next in rule order).
- **Repository** — two sends; `findAll()` contains both message IDs.
- **Cost report** — after one success, provider/country totals, volume, and success rate match expected values.
