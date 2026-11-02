# Developer Guide — Running the E2E SMS Test Scenario

## Prerequisites

- Java 25 (Temurin or equivalent)
- No external services required — all storage is in-memory

## 1. Start the application

From the project root:

```bash
./gradlew bootRun
```

The application listens on **http://localhost:8080**.

---

## 2. Run the happy path (MSG-VN-001)

### Step 1+2 — Submit SMS and trigger routing

```bash
curl -X POST http://localhost:8080/api/v1/sendSms \
  -H "Content-Type: application/json" \
  -d '{
    "messageId":   "MSG-VN-001",
    "country":     "VN",
    "phoneNumber": "+84912345678",
    "message":     "Your OTP is 482910"
  }'
```

**Expected response:**

```
send success
```

**What happens internally:**
- Phone prefix `+8491` resolves to carrier `Viettel`
- Routing rule `VN + Viettel → Vonage` is applied, `estimatedCost = 0.030`
- Message state transitions: `NewMessageRequested → RoutePlanCalculated → SentToProvider`

---

### Step 3 — Simulate provider callback: Queue

```bash
curl "http://localhost:8080/api/v1/callback-simulation?messageId=MSG-VN-001&state=Queue"
```

**Expected response:** HTTP 200

**State after:** `SentToProvider → Queued`

---

### Step 4 — Simulate provider callback: Send-to-carrier

```bash
curl "http://localhost:8080/api/v1/callback-simulation?messageId=MSG-VN-001&state=Send-to-carrier"
```

**Expected response:** HTTP 200

**State after:** `Queued → SentToCarrier`

---

### Step 5 — Simulate provider callback: Send-success

```bash
curl "http://localhost:8080/api/v1/callback-simulation?messageId=MSG-VN-001&state=Send-success&actualCost=0.027"
```

**Expected response:** HTTP 200

**State after:** `SentToCarrier → SentSuccessfully`  
`actualCost = 0.027` is stored. `estimatedCost = 0.030` remains for variance tracking.

---

### Send 3 additional VN messages for aggregation

To see meaningful multi-provider data in the reports below, send 3 more messages via Vinaphone (routes to Twilio) and run their full callback lifecycle before querying.

Considering for VN `actualCost = 0.027` and `estimatedCost = 0.030`

```bash
//Message 1
curl -X POST http://localhost:8080/api/v1/sendSms \
  -H "Content-Type: application/json" \
  -d '{
    "messageId":   "MSG-VN-002",
    "country":     "VN",
    "phoneNumber": "+84882345670",
    "message":     "test message 1"
  }'
curl "http://localhost:8080/api/v1/callback-simulation?messageId=MSG-VN-002&state=Queue"
curl "http://localhost:8080/api/v1/callback-simulation?messageId=MSG-VN-002&state=Send-to-carrier"
curl "http://localhost:8080/api/v1/callback-simulation?messageId=MSG-VN-002&state=Send-success&actualCost=0.013"

//Message 2
curl -X POST http://localhost:8080/api/v1/sendSms \
  -H "Content-Type: application/json" \
  -d '{
    "messageId":   "MSG-VN-003",
    "country":     "VN",
    "phoneNumber": "+84882345671",
    "message":     "test message 2"
  }'
curl "http://localhost:8080/api/v1/callback-simulation?messageId=MSG-VN-003&state=Queue"
curl "http://localhost:8080/api/v1/callback-simulation?messageId=MSG-VN-003&state=Send-to-carrier"
curl "http://localhost:8080/api/v1/callback-simulation?messageId=MSG-VN-003&state=Send-success&actualCost=0.013"

//Message 3
curl -X POST http://localhost:8080/api/v1/sendSms \
  -H "Content-Type: application/json" \
  -d '{
    "messageId":   "MSG-VN-004",
    "country":     "VN",
    "phoneNumber": "+84882345672",
    "message":     "test message 3"
  }'
curl "http://localhost:8080/api/v1/callback-simulation?messageId=MSG-VN-004&state=Queue"
curl "http://localhost:8080/api/v1/callback-simulation?messageId=MSG-VN-004&state=Send-to-carrier"
curl "http://localhost:8080/api/v1/callback-simulation?messageId=MSG-VN-004&state=Send-success&actualCost=0.014"
```


---

### Step 6 — Query cost report by country

```bash
curl http://localhost:8080/api/v1/costPerCountry
```

**Expected response** (4 messages: 1 Vonage + 3 Twilio):

```json
[
  {
    "country": "VN",
    "estimatedCost": 0.075,
    "actualCost": 0.067
  }
]
```

- `estimatedCost = 0.075` → 0.030 (Vonage) + 3 × 0.015 (Twilio)
- `actualCost = 0.067` → 0.027 (MSG-VN-001) + 0.013 + 0.013 + 0.014

---

### Step 7 — Query cost report by provider

```bash
curl http://localhost:8080/api/v1/costPerProvider
```

**Expected response:**

```json
[
  {
    "providerId": "Vonage",
    "estimatedCost": 0.03,
    "actualCost": 0.027
  },
  {
    "providerId": "Twilio",
    "estimatedCost": 0.045,
    "actualCost": 0.04
  }
]
```

- Provider totals sum to the country totals in Step 6
- `Twilio.estimatedCost = 0.045` → 3 × 0.015

---

### Step 8 — Query SMS volume by provider

```bash
curl http://localhost:8080/api/v1/smsVolumePerProvider
```

**Expected response:**

```json
[
  {
    "providerId": "Vonage",
    "messageCount": 1
  },
  {
    "providerId": "Twilio",
    "messageCount": 3
  }
]
```

- `messageCount` reflects the number of routed messages, regardless of final delivery status

### Step 9 — Query delivery rate by provider

To see a meaningful failure rate, send MSG-VN-005 (Viettel → Vonage) and simulate a carrier rejection before querying.

```bash
curl -X POST http://localhost:8080/api/v1/sendSms \
  -H "Content-Type: application/json" \
  -d '{
    "messageId":   "MSG-VN-005",
    "country":     "VN",
    "phoneNumber": "+84912345679",
    "message":     "test failure"
  }'

curl "http://localhost:8080/api/v1/callback-simulation?messageId=MSG-VN-005&state=Queue"
curl "http://localhost:8080/api/v1/callback-simulation?messageId=MSG-VN-005&state=Carrier-rejected"
```

```bash
curl http://localhost:8080/api/v1/deliveryRatePerProvider
```

**Expected response** (MSG-VN-001 succeeded + MSG-VN-005 rejected = 2 Vonage; 3 Twilio all succeeded):

```json
[
  {
    "providerId": "Vonage",
    "sent": 2,
    "succeeded": 1,
    "failed": 1,
    "successRate": 0.5,
    "failureRate": 0.5
  },
  {
    "providerId": "Twilio",
    "sent": 3,
    "succeeded": 3,
    "failed": 0,
    "successRate": 1.0,
    "failureRate": 0.0
  }
]
```

- `Vonage.successRate = 0.5` — 1 of 2 messages reached the end user
- `Twilio.successRate = 1.0` — all 3 Vinaphone messages delivered

---

## 3. Run the negative cases

Each negative case requires a fresh message. Start the app fresh or use a new `messageId` per test.

### Case 1 — Skip Queue (Send-to-provider → Send-to-carrier directly)

```bash
curl -X POST http://localhost:8080/api/v1/sendSms \
  -H "Content-Type: application/json" \
  -d '{"messageId":"MSG-NEG-001","country":"VN","phoneNumber":"+84912345678","message":"test"}'

curl "http://localhost:8080/api/v1/callback-simulation?messageId=MSG-NEG-001&state=Send-to-carrier"
```

**Expected:** HTTP 500 — transition rejected (message is in `SentToProvider`, not `Queued`)

---

### Case 2 — Jump to success (New → Send-success)

```bash
curl -X POST http://localhost:8080/api/v1/sendSms \
  -H "Content-Type: application/json" \
  -d '{"messageId":"MSG-NEG-002","country":"VN","phoneNumber":"+84912345678","message":"test"}'

curl "http://localhost:8080/api/v1/callback-simulation?messageId=MSG-NEG-002&state=Send-success&actualCost=0.01"
```

**Expected:** HTTP 500 — transition rejected (message is in `SentToProvider`, not `SentToCarrier`)

---

### Case 3 — Revert state (Send-success → Queue)

```bash
curl -X POST http://localhost:8080/api/v1/sendSms \
  -H "Content-Type: application/json" \
  -d '{"messageId":"MSG-NEG-003","country":"VN","phoneNumber":"+84912345678","message":"test"}'

curl "http://localhost:8080/api/v1/callback-simulation?messageId=MSG-NEG-003&state=Queue"
curl "http://localhost:8080/api/v1/callback-simulation?messageId=MSG-NEG-003&state=Send-to-carrier"
curl "http://localhost:8080/api/v1/callback-simulation?messageId=MSG-NEG-003&state=Send-success&actualCost=0.01"

curl "http://localhost:8080/api/v1/callback-simulation?messageId=MSG-NEG-003&state=Queue"
```

**Expected:** Last request returns HTTP 500 — backward transition rejected

---

### Case 4 — Unknown messageId

```bash
curl "http://localhost:8080/api/v1/callback-simulation?messageId=MSG-UNKNOWN&state=Queue"
```

**Expected:** HTTP 500 — `Message not found: MSG-UNKNOWN`

---

### Case 5 — Duplicate callback (idempotency)

Using `MSG-NEG-003` from Case 3 which is already in `SentSuccessfully`:

```bash
curl "http://localhost:8080/api/v1/callback-simulation?messageId=MSG-NEG-003&state=Send-success&actualCost=0.01"
```

**Expected:** HTTP 200 — callback silently ignored, no duplicate event recorded

---

## 4. Available report endpoints

| Endpoint | Description |
|---|---|
| `GET /api/v1/costPerCountry` | Estimated and actual cost aggregated by country |
| `GET /api/v1/costPerProvider` | Estimated and actual cost aggregated by provider |
| `GET /api/v1/smsVolumePerProvider` | Message count aggregated by provider |
| `GET /api/v1/deliveryRatePerProvider` | Sent, succeeded, failed counts and success/failure rates by provider |

All reports are built by replaying the event store — no separate read model is maintained.

---

## 5. Supported phone prefixes (carrier simulation)

| Prefix | Carrier | Country |
|---|---|---|
| +8491 / +8496 / +8497 | Viettel | VN |
| +8490 / +8493 | Mobifone | VN |
| +8488 / +8494 | Vinaphone | VN |
| +6681 / +6682 | AIS | TH |
| +6683 / +6684 | DTAC | TH |
| +6591 / +6592 | Singtel | SG |
| +6581 / +6582 | StarHub | SG |
| +6391 / +6392 | Globe | PH |
| +6394 / +6395 | Smart | PH |
| +6389 | DITO | PH |

## 6. Routing rules (current agreements)

| Country | Carrier | Provider | Estimated cost |
|---|---|---|---|
| VN | Viettel | Vonage | $0.030 |
| VN | Mobifone | Infobip | $0.019 |
| VN | Vinaphone | Twilio | $0.015 |
| TH | AIS | Infobip | $0.025 |
| TH | DTAC | AWSSNS | $0.022 |
| SG | Singtel | Twilio | $0.035 |
| SG | StarHub | Telnyx | $0.030 |
| PH | Globe | MessageBird | $0.028 |
| PH | Smart | Sinch | $0.026 |
| PH | DITO | MessageBird | $0.028 |
