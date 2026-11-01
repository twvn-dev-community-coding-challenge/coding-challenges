-- A ride can end before dispatch has opened its request (its RideRequested was dead-lettered or delayed). Dispatch
-- then records the ride as already ended, without a pickup, so a late RideRequested conflicts on the primary key and
-- the ride is never offered. Every request that can still be matched keeps its pickup.
ALTER TABLE dispatch_requests
    ALTER COLUMN pickup_latitude DROP NOT NULL,
    ALTER COLUMN pickup_longitude DROP NOT NULL;

ALTER TABLE dispatch_requests ADD CONSTRAINT dispatch_requests_open_has_pickup CHECK (
    status IN ('CANCELLED', 'COMPLETED')
    OR (pickup_latitude IS NOT NULL AND pickup_longitude IS NOT NULL));
