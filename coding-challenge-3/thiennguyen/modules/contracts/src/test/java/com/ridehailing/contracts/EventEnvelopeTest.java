package com.ridehailing.contracts;

import static org.assertj.core.api.Assertions.assertThat;

import java.time.Instant;
import java.util.Map;
import org.junit.jupiter.api.Test;

class EventEnvelopeTest {

    @Test
    void aCreatedEventHasAFreshIdTheWireNameAndTheCurrentSchemaVersion() {
        Instant occurredAt = Instant.parse("2026-10-26T10:00:00Z");
        Map<String, Object> payload = Map.of("rideId", "R1");

        EventEnvelope first = EventEnvelope.create(EventType.RIDE_CANCELLED, "R1", null, occurredAt, payload);
        EventEnvelope second = EventEnvelope.create(EventType.RIDE_CANCELLED, "R1", null, occurredAt, payload);

        assertThat(first.eventId()).isNotNull().isNotEqualTo(second.eventId());
        assertThat(first.eventType()).isEqualTo(EventType.RIDE_CANCELLED.wireName());
        assertThat(first.rideId()).isEqualTo("R1");
        assertThat(first.driverId()).isNull();
        assertThat(first.occurredAt()).isEqualTo(occurredAt);
        assertThat(first.schemaVersion()).isEqualTo(EventEnvelope.CURRENT_SCHEMA_VERSION);
        assertThat(first.payload()).isEqualTo(payload);
    }
}
