# Product & engineering backlog

Structured backlog cards for **`gondo`**. Priority bands match [system-vs-program-requirements.md](../system-vs-program-requirements.md).

| Priority | Meaning |
|----------|---------|
| **P0** | Submission / reviewer experience — do before merge when possible |
| **P1** | Core challenge fit — correctness, tests, cost semantics |
| **P2** | Enhancements — often implemented or partially implemented; track hardening |
| **Stretch** | Explicitly optional — persistence, demos, bonuses |

**Status legend:** **Done** · **Done (minor gaps)** · **Mostly done** · **Partial** · **Open** — last reviewed against the repo in this folder.

## Index

| ID | Status | File | Summary |
|----|--------|------|---------|
| P0-01 | **Done (minor gaps)** | [p0-submission-readme.md](p0-submission-readme.md) | Template + links; optional **AI tools** line |
| P0-02 | **Done** | [p0-judge-quickstart-limitations.md](p0-judge-quickstart-limitations.md) | Quickstart + limitations in README |
| P1-01 | **Done** | [p1-cost-story.md](p1-cost-story.md) | Charging estimate/actual + tests |
| P1-02 | **Mostly done** | [p1-resilience-tests.md](p1-resilience-tests.md) | 503 / 409 / charging errors covered |
| P1-03 | **Done** | [p1-routing-seeds-user-story-3.md](p1-routing-seeds-user-story-3.md) | Story 3 vs seeds in routing doc |
| P2-01 | **Partial** | [p2-otp-service-hardening.md](p2-otp-service-hardening.md) | Service shipped; prod hardening open |
| P2-02 | **Partial** | [p2-openapi-platform-follow-ups.md](p2-openapi-platform-follow-ups.md) | Export done; CI/portal/SDK open |
| ST-01 | **Open** | [stretch-notification-persistence.md](stretch-notification-persistence.md) | Durable notifications |
| ST-02 | **Open** | [stretch-demo-script.md](stretch-demo-script.md) | Scripted demo |
| ST-03 | **Open** | [stretch-program-bonuses.md](stretch-program-bonuses.md) | Process bonuses |

## Related docs

- [system-vs-program-requirements.md](../system-vs-program-requirements.md) — master mapping & table backlog
- [platform-sms-openapi.md](../platform-sms-openapi.md) — OpenAPI export story
- [routing-vs-challenge-brief.md](../routing-vs-challenge-brief.md) — User Story 3
