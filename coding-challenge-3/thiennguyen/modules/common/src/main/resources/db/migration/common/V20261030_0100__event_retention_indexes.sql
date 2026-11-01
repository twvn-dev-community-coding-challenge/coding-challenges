-- Lets EventRetention find expired rows without scanning either table (ADR-005).

CREATE INDEX outbox_published_at_idx ON outbox (published_at) WHERE published_at IS NOT NULL;
CREATE INDEX processed_events_processed_at_idx ON processed_events (processed_at);
