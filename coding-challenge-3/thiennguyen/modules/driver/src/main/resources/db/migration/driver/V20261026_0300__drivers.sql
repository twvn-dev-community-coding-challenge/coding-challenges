-- availability and location are owned here. assignment/current_ride_id are a versioned copy of dispatch's decision
-- (ADR-002), kept so the API can show OFFERED/BUSY and refuse to go offline mid-offer or mid-trip (AC-02).
CREATE TABLE drivers (
    id                  VARCHAR(64) PRIMARY KEY,
    availability        VARCHAR(16) NOT NULL,
    latitude            DOUBLE PRECISION,
    longitude           DOUBLE PRECISION,
    location_updated_at TIMESTAMPTZ,
    assignment          VARCHAR(16) NOT NULL DEFAULT 'NONE',
    -- Ride owned by the ride module; no foreign key, so this table can move to its own database.
    current_ride_id     UUID,
    assignment_version  BIGINT      NOT NULL DEFAULT 0,
    version             BIGINT      NOT NULL DEFAULT 0,
    updated_at          TIMESTAMPTZ NOT NULL,
    CONSTRAINT drivers_availability_valid CHECK (availability IN ('OFFLINE', 'AVAILABLE')),
    CONSTRAINT drivers_assignment_valid CHECK (assignment IN ('NONE', 'OFFERED', 'BUSY')),
    CONSTRAINT drivers_location_valid CHECK (
        (latitude IS NULL AND longitude IS NULL)
        OR (latitude BETWEEN -90 AND 90 AND longitude BETWEEN -180 AND 180)),
    CONSTRAINT drivers_ride_matches_assignment CHECK ((assignment <> 'NONE') = (current_ride_id IS NOT NULL))
);
