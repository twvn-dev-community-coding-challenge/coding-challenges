# ADR-004 — Driver matching: nearest available driver within a radius

- **Status:** Accepted
- **Covers:** AC-03, AC-04, AC-07

## Context

The brief requires selecting the nearest `AVAILABLE` driver within a configurable radius (default 2 km), using each driver's latest known location, breaking distance ties by the lexicographically smallest `driverId`, and permanently excluding drivers who rejected (or let expire) an offer for the same ride. Scale for the implementation is thousands of drivers; the design must show how it evolves to 1 M.

## Decision

1. **Two-stage selection.**
   - *Pre-filter in SQL* on dispatch's supply copy (`driver_supply`): `online AND assignment = 'NONE'`, location inside a lat/lng **bounding box** derived from the radius (`BoundingBox`, which handles the poles and the antimeridian). Served by the partial index `driver_supply_free_location_idx (latitude, longitude) WHERE online AND assignment = 'NONE'`.
   - *Rank in Java:* `NearestDriverMatcher`, a pure function. It drops drivers excluded for this ride, computes **Haversine** distance (`GeoDistance`, Earth mean radius 6371.0088 km), drops candidates with distance `> radius`, and sorts by `(distanceKm, driverId)`.
2. The matcher returns an **ordered candidate list**, not a single driver. Dispatch walks the list and atomically claims the first driver still free ([ADR-006](ADR-006-concurrency.md)); losing a race falls through to the next candidate instead of failing the ride.
3. Radius is `dispatch.max-radius-km` (default `2.0`); the boundary is inclusive (`distance ≤ radius`).
4. Exclusions live in `ride_driver_exclusions(ride_id, driver_id)` and are written in the same transaction as the reject or expiry, so they are never lost. A driver whose offer expired is excluded exactly like one who rejected.
5. **Latest known location.** The driver module overwrites its `drivers` row and publishes `DriverLocationUpdated`; dispatch applies it to `driver_supply` only if its version is newer ([ADR-002](ADR-002-decoupled-contexts.md)). Matching reads `driver_supply` inside its own transaction, so it uses the latest location dispatch has consumed, normally a few milliseconds behind the driver's last report.

## Consequences

- The matcher is deterministic and has no I/O, so it is unit-tested exhaustively (`NearestDriverMatcherTest`: nearest, outside radius, offline, busy, offered, none, tie-break, exclusions, inclusive boundary; `BoundingBoxTest`: poles and antimeridian). `AC04MatchingAcceptanceIT` and `AC07RematchingAcceptanceIT` cover the same rules end to end.
- Straight-line distance ignores roads and ETA (accepted limitation).
- Greedy per-ride matching can be globally sub-optimal when many rides compete in one area.
- Bounding-box pre-filter keeps the Java ranking set small; cost is O(k log k) for k drivers in the box.

## Alternatives considered

| Option | Why not now |
|---|---|
| PostGIS `ST_DWithin` + KNN `<->` | Best fit at medium scale and the first evolution step; adds an extension and image for little gain at seed-data size. |
| Redis `GEOSEARCH` | Designed, **not built**: an optional candidate source that only proposes candidates while the Postgres claim still decides ([ADR-008](ADR-008-caching.md)). Postgres is enough at this scale. |
| H3 / geohash cells | Needed for partitioning dispatch by area at 1 M drivers; premature here. |
| Batch / global assignment (Hungarian, min-cost flow) | Better utilisation in dense markets, but the brief explicitly asks for nearest-driver. |
