-- Owned by the common module: the transactional outbox and consumer de-duplication (ADR-007, ADR-005).
-- When modules become services, each service gets its own copy of both tables.

CREATE TABLE outbox (
    id           BIGSERIAL PRIMARY KEY,
    event_id     UUID         NOT NULL UNIQUE,
    topic        VARCHAR(100) NOT NULL,
    message_key  VARCHAR(64)  NOT NULL,
    aggregate_id VARCHAR(64)  NOT NULL,
    event_type   VARCHAR(64)  NOT NULL,
    payload      JSONB        NOT NULL,
    created_at   TIMESTAMPTZ  NOT NULL,
    published_at TIMESTAMPTZ
);

CREATE INDEX outbox_unpublished_idx ON outbox (id) WHERE published_at IS NULL;
CREATE INDEX outbox_aggregate_idx ON outbox (aggregate_id, event_type);

CREATE TABLE processed_events (
    consumer     VARCHAR(64) NOT NULL,
    event_id     UUID        NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (consumer, event_id)
);
