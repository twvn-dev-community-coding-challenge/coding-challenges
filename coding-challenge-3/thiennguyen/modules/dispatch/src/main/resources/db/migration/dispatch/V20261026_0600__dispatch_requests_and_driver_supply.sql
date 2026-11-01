-- ADR-002: dispatch owns everything it decides with, so a matching attempt, an acceptance or a release is one local
-- transaction. ride_id and driver_id refer to the ride and driver modules; no foreign keys cross that boundary.

-- One row per ride dispatch has been asked to serve, created from RideRequested (pickup travels in the event).
CREATE TABLE dispatch_requests (
    ride_id          UUID PRIMARY KEY,
    pickup_latitude  DOUBLE PRECISION NOT NULL,
    pickup_longitude DOUBLE PRECISION NOT NULL,
    status           VARCHAR(16)      NOT NULL,
    driver_id        VARCHAR(64),
    version          BIGINT           NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ      NOT NULL,
    updated_at       TIMESTAMPTZ      NOT NULL,
    CONSTRAINT dispatch_requests_status_valid CHECK (
        status IN ('WAITING', 'OFFERED', 'ACCEPTED', 'CANCELLED', 'COMPLETED')),
    CONSTRAINT dispatch_requests_active_has_driver CHECK (
        status NOT IN ('OFFERED', 'ACCEPTED') OR driver_id IS NOT NULL),
    CONSTRAINT dispatch_requests_waiting_has_no_driver CHECK (status <> 'WAITING' OR driver_id IS NULL)
);

-- ADR-006: one driver never holds two active rides, even if application code regresses.
CREATE UNIQUE INDEX dispatch_requests_active_driver_uq ON dispatch_requests (driver_id)
    WHERE status IN ('OFFERED', 'ACCEPTED');
CREATE INDEX dispatch_requests_waiting_idx ON dispatch_requests (updated_at) WHERE status = 'WAITING';

-- Dispatch's supply: whether each driver is online and where (a copy fed by driver-events, versioned per fact), and
-- whether dispatch holds the driver (owned here: the claim is a compare-and-set on this row).
CREATE TABLE driver_supply (
    driver_id            VARCHAR(64) PRIMARY KEY,
    online               BOOLEAN     NOT NULL DEFAULT FALSE,
    latitude             DOUBLE PRECISION,
    longitude            DOUBLE PRECISION,
    availability_version BIGINT      NOT NULL DEFAULT 0,
    location_version     BIGINT      NOT NULL DEFAULT 0,
    assignment           VARCHAR(16) NOT NULL DEFAULT 'NONE',
    current_ride_id      UUID,
    version              BIGINT      NOT NULL DEFAULT 0,
    updated_at           TIMESTAMPTZ NOT NULL,
    CONSTRAINT driver_supply_assignment_valid CHECK (assignment IN ('NONE', 'OFFERED', 'BUSY')),
    CONSTRAINT driver_supply_location_valid CHECK (
        (latitude IS NULL AND longitude IS NULL)
        OR (latitude BETWEEN -90 AND 90 AND longitude BETWEEN -180 AND 180)),
    CONSTRAINT driver_supply_ride_matches_assignment CHECK ((assignment <> 'NONE') = (current_ride_id IS NOT NULL))
);

-- A ride holds at most one driver at a time.
CREATE UNIQUE INDEX driver_supply_current_ride_uq ON driver_supply (current_ride_id) WHERE current_ride_id IS NOT NULL;
CREATE INDEX driver_supply_free_location_idx ON driver_supply (latitude, longitude)
    WHERE online AND assignment = 'NONE';
