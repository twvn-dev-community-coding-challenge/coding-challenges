-- ride_id and driver_id belong to the ride and driver modules; no foreign keys, so dispatch can own its database.
CREATE TABLE ride_offers (
    id           UUID PRIMARY KEY,
    ride_id      UUID             NOT NULL,
    driver_id    VARCHAR(64)      NOT NULL,
    status       VARCHAR(16)      NOT NULL,
    distance_km  DOUBLE PRECISION NOT NULL,
    created_at   TIMESTAMPTZ      NOT NULL,
    expires_at   TIMESTAMPTZ      NOT NULL,
    responded_at TIMESTAMPTZ,
    CONSTRAINT ride_offers_status_valid CHECK (status IN ('PENDING', 'ACCEPTED', 'REJECTED', 'EXPIRED', 'CANCELLED'))
);

CREATE UNIQUE INDEX ride_offers_pending_driver_uq ON ride_offers (driver_id) WHERE status = 'PENDING';
CREATE UNIQUE INDEX ride_offers_pending_ride_uq ON ride_offers (ride_id) WHERE status = 'PENDING';
CREATE INDEX ride_offers_pending_expiry_idx ON ride_offers (expires_at) WHERE status = 'PENDING';
CREATE INDEX ride_offers_ride_idx ON ride_offers (ride_id);

CREATE TABLE ride_driver_exclusions (
    ride_id    UUID        NOT NULL,
    driver_id  VARCHAR(64) NOT NULL,
    reason     VARCHAR(16) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (ride_id, driver_id)
);
