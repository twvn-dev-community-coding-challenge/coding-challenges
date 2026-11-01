-- passenger_id and assigned_driver_id reference other modules and are validated by the application, not by foreign
-- keys, so this module's tables can move to their own database when ride becomes a service.
CREATE TABLE rides (
    id                    UUID PRIMARY KEY,
    passenger_id          VARCHAR(64)      NOT NULL,
    pickup_latitude       DOUBLE PRECISION NOT NULL,
    pickup_longitude      DOUBLE PRECISION NOT NULL,
    destination_latitude  DOUBLE PRECISION NOT NULL,
    destination_longitude DOUBLE PRECISION NOT NULL,
    status                VARCHAR(20)      NOT NULL,
    assigned_driver_id    VARCHAR(64),
    version               BIGINT           NOT NULL DEFAULT 0,
    created_at            TIMESTAMPTZ      NOT NULL,
    updated_at            TIMESTAMPTZ      NOT NULL,
    CONSTRAINT rides_status_valid CHECK (
        status IN ('REQUESTED', 'MATCHING', 'DRIVER_ASSIGNED', 'DRIVER_ACCEPTED', 'CANCELLED', 'COMPLETED')),
    CONSTRAINT rides_assigned_requires_driver CHECK (
        status NOT IN ('DRIVER_ASSIGNED', 'DRIVER_ACCEPTED') OR assigned_driver_id IS NOT NULL),
    CONSTRAINT rides_unmatched_has_no_driver CHECK (
        status NOT IN ('REQUESTED', 'MATCHING') OR assigned_driver_id IS NULL),
    CONSTRAINT rides_pickup_valid CHECK (pickup_latitude BETWEEN -90 AND 90 AND pickup_longitude BETWEEN -180 AND 180),
    CONSTRAINT rides_destination_valid CHECK (
        destination_latitude BETWEEN -90 AND 90 AND destination_longitude BETWEEN -180 AND 180)
);

-- assigned_driver_id mirrors dispatch's offer state (ADR-002). The one-active-ride-per-driver guarantee is enforced
-- where the assignment is decided, in dispatch's own tables; a projection that updates rides on different partitions
-- independently must not carry it, or a correct but out-of-order delivery would be rejected.
CREATE INDEX rides_assigned_driver_idx ON rides (assigned_driver_id) WHERE assigned_driver_id IS NOT NULL;
CREATE INDEX rides_passenger_idx ON rides (passenger_id);

CREATE TABLE idempotency_keys (
    key          VARCHAR(128) PRIMARY KEY,
    request_hash CHAR(64)     NOT NULL,
    ride_id      UUID         NOT NULL REFERENCES rides (id) DEFERRABLE INITIALLY DEFERRED,
    created_at   TIMESTAMPTZ  NOT NULL
);

CREATE TABLE ride_state_history (
    id          BIGSERIAL PRIMARY KEY,
    ride_id     UUID         NOT NULL REFERENCES rides (id),
    from_status VARCHAR(20),
    to_status   VARCHAR(20)  NOT NULL,
    driver_id   VARCHAR(64),
    reason      VARCHAR(100) NOT NULL,
    occurred_at TIMESTAMPTZ  NOT NULL
);

CREATE INDEX ride_state_history_ride_idx ON ride_state_history (ride_id, id);
