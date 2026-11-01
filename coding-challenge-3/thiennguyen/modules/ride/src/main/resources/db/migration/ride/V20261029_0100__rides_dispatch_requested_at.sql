-- When RideRequested was last published for a ride. A ride still REQUESTED long after that has not been picked up by
-- dispatch (its event was dead-lettered or lost), so UndispatchedRideSweeper publishes RideRequested again.
ALTER TABLE rides ADD COLUMN dispatch_requested_at TIMESTAMPTZ;
UPDATE rides SET dispatch_requested_at = created_at;
ALTER TABLE rides ALTER COLUMN dispatch_requested_at SET NOT NULL;

CREATE INDEX rides_undispatched_idx ON rides (dispatch_requested_at) WHERE status = 'REQUESTED';
