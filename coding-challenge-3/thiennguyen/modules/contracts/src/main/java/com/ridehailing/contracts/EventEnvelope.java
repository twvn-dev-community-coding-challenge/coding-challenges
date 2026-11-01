package com.ridehailing.contracts;

import java.time.Instant;
import java.util.Map;
import java.util.UUID;

/**
 * Every message on every topic. {@code eventId} identifies this event for consumer de-duplication; {@code rideId}
 * and {@code driverId} are correlation identifiers (either may be null); {@code payload} is the JSON form of the
 * {@link EventType#payloadType()} record for {@code schemaVersion}.
 */
public record EventEnvelope(
        UUID eventId,
        String eventType,
        String rideId,
        String driverId,
        Instant occurredAt,
        int schemaVersion,
        Map<String, Object> payload) {

    public static final int CURRENT_SCHEMA_VERSION = 1;

    /** A new event of {@code type}, with a fresh {@code eventId} and the current schema version. */
    public static EventEnvelope create(EventType type, String rideId, String driverId, Instant occurredAt,
            Map<String, Object> payload) {
        return new EventEnvelope(UUID.randomUUID(), type.wireName(), rideId, driverId, occurredAt,
                CURRENT_SCHEMA_VERSION, payload);
    }
}
