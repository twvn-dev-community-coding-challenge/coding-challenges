package com.ridehailing.common.events;

import com.ridehailing.contracts.EventType;
import java.util.UUID;

/**
 * Records an event inside the caller's transaction, so the event exists if and only if the state change that caused it
 * committed; it is published later, to the event type's own topic. Each call must run in a transaction.
 */
public interface OutboxWriter {

    /** Keyed by ride, so every event about one ride lands on one partition in order. Returns the event id. */
    UUID appendForRide(EventType type, UUID rideId, String driverId, Object payload);

    /** Keyed by driver, so every event about one driver lands on one partition in order. Returns the event id. */
    UUID appendForDriver(EventType type, String driverId, UUID rideId, Object payload);
}
