# Light Nguyen - Coding Challenge 3: Ride-Hailing Dispatch Service

## Team Members

- Light Nguyen (backend developer)

## How to Run

Requirements: Docker with Docker Compose. From the repository root, start the
Ride Management API, Driver Management API, and RabbitMQ:

```bash
cd coding-challenge-3/lightnguyen
docker compose up --build
```

The APIs are available at:

- Ride Management: `http://localhost:5080`
- Driver Management: `http://localhost:5081`
- RabbitMQ management UI: `http://localhost:15672` (user `challenge`, password
  `challenge`)

To stop the services while retaining their database and broker data, run
`docker compose down` from the same directory.

Run the test projects from `coding-challenge-3/lightnguyen` with:

```bash
dotnet test CodingChallenge.DriverManagement.Tests/CodingChallenge.DriverManagement.Tests.csproj
dotnet test CodingChallenge.RideManagement.Tests/CodingChallenge.RideManagement.Tests.csproj
```

### Try the dispatch flow

Create a driver, then set the driver's location and availability:

```http
POST http://localhost:5081/api/Driver
Content-Type: application/json

{"name":"Test driver"}
```

Use the returned `id` for `driverId` in the following requests:

```http
PUT http://localhost:5081/api/drivers/{{driverId}}/location
Content-Type: application/json

{"latitude":37.7749,"longitude":-122.4194}
```

```http
PUT http://localhost:5081/api/drivers/{{driverId}}/availability
Content-Type: application/json

true
```

Request a ride with a fresh `Idempotency-Key` for each new ride. `passengerId`
can be any UUID; passengers are not managed by a separate API in this flow.

```http
POST http://localhost:5080/Rides
Idempotency-Key: 00000000-0000-0000-0000-000000000101
Content-Type: application/json

{
  "passengerId": "00000000-0000-0000-0000-000000000201",
  "pickup": {"latitude":37.7749,"longtitude":-122.4194},
  "destination": {"latitude":37.7849,"longtitude":-122.4094}
}
```

Save the returned ride `id` as `rideId`. The create response does not wait for
driver matching, so poll
`GET http://localhost:5080/Rides/{{rideId}}` until `assignedDriverId` matches
the driver's ID. The driver's pending offer is available at
`GET http://localhost:5081/api/drivers/{{driverId}}/offers`.

Accept an offer with
`POST http://localhost:5080/Rides/{{rideId}}/offers/{{driverId}}/accept`.
Reject it with
`POST http://localhost:5080/Rides/{{rideId}}/offers/{{driverId}}/reject`.
Both actions return `202 Accepted`; poll the ride and driver-offers endpoints
to observe the result. Acceptance advances the ride to `DRIVER_ACCEPTED`.
Rejection marks the offer `REJECTED` and dispatches to the next closest
available driver, so register and position a second driver before trying it.

The shared ride model currently spells the coordinate property `longtitude`;
driver-location requests use `longitude`. API enum statuses are serialized as
strings, such as `DRIVER_ASSIGNED` and `PENDING`.

## Approach

The solution separates ride handling from driver dispatch. Ride Management
persists a new ride and publishes a persistent `RideCreated` message to
RabbitMQ. Driver Management consumes the event, finds the nearest available
driver, creates an offer, and publishes `DriverMatched`; Ride Management then
updates the ride. The API waits for the broker to confirm publication, not for
matching to finish.

Delivery is at-least-once. The consumer reuses an existing active offer for a
ride instead of creating a duplicate, and ride state handlers ignore duplicate
or stale matching events. On processing failure, the consumer rejects the
message without requeueing; RabbitMQ routes it to the queue's durable `.dead`
queue for investigation. There is no automatic retry or dead-letter replay
policy.

## Design Decisions

- **Tech stack:** ASP.NET Core and .NET 10 provide the APIs and hosted
  consumers; EF Core with SQLite keeps local persistence simple; RabbitMQ
  decouples dispatch work from ride creation; Docker Compose starts the
  services and broker together. xUnit is used for automated tests.
- **Trade-offs:** Asynchronous dispatch keeps the request path short and
  services independent, but clients must poll for the match. SQLite lowers
  setup overhead but may need reconsideration for a larger production
  deployment. Failed messages are retained for inspection, but recovery
  currently requires operator investigation and republishing.
- **Patterns:** Service separation, event-driven messaging, explicit ride
  lifecycle transitions, and idempotent request/event handling. Persisted
  business state and database constraints provide duplicate and concurrent
  assignment protection; there is no separate message-ID deduplication store.

## Challenges Faced

The dispatch result is not available when the ride-create request returns, so
the API and test flow need to account for eventual completion. At-least-once
delivery also means the same event may arrive repeatedly. The design handles
these cases by publishing durable events, reusing active offers, guarding
lifecycle transitions, and making ride creation retryable with the same
`Idempotency-Key`. Broker outages can still cause publication to fail; clients
can retry the request to republish the event for the already-persisted ride.

## What We Learned

- How to keep asynchronous work out of an HTTP request while coordinating
  services with RabbitMQ.
- Why message consumers and domain transitions need to tolerate duplicate or
  stale events.
- How database constraints and persisted state can help protect business
  invariants under retries and concurrent requests.

## With More Time, We Would...

- Add a deliberate retry/backoff policy and an operational dead-letter replay
  workflow.
- Add end-to-end integration tests that exercise the APIs and RabbitMQ
  together, including broker failure and recovery.
- Improve operational visibility with structured tracing and metrics across
  the ride and dispatch event flow.

## AI Tools Used (if any)

- GitHub Copilot in VS Code — used to help update and organize this README.
