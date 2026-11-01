package com.ridehailing.common.events;

import com.ridehailing.common.jdbc.Timestamps;
import com.ridehailing.common.logging.LogContext;
import com.ridehailing.contracts.EventEnvelope;
import com.ridehailing.contracts.EventType;

import java.time.Clock;
import java.util.UUID;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.jdbc.core.simple.JdbcClient;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Propagation;
import org.springframework.transaction.annotation.Transactional;

/** {@link OutboxWriter} on the {@code outbox} table, which {@link OutboxRelay} publishes from. */
@Component
@RequiredArgsConstructor
@Slf4j
public class JdbcOutboxWriter implements OutboxWriter {

    private final JdbcClient jdbc;
    private final EventCodec codec;
    private final Clock clock;

    @Override
    @Transactional(propagation = Propagation.MANDATORY)
    public UUID appendForRide(EventType type, UUID rideId, String driverId, Object payload) {
        return append(type, rideId.toString(), rideId.toString(), driverId, payload);
    }

    @Override
    @Transactional(propagation = Propagation.MANDATORY)
    public UUID appendForDriver(EventType type, String driverId, UUID rideId, Object payload) {
        return append(type, driverId, rideId == null ? null : rideId.toString(), driverId, payload);
    }

    private UUID append(EventType type, String messageKey, String rideId, String driverId, Object payload) {
        if (!type.payloadType().isInstance(payload)) {
            throw new IllegalArgumentException(
                    "%s carries %s, not %s".formatted(type, type.payloadType().getSimpleName(), payload.getClass()));
        }
        EventEnvelope envelope =
                EventEnvelope.create(type, rideId, driverId, clock.instant(), codec.toPayload(payload));
        jdbc.sql("""
                        INSERT INTO outbox (event_id, topic, message_key, aggregate_id, event_type, payload, created_at)
                        VALUES (:eventId, :topic, :messageKey, :aggregateId, :eventType, CAST(:payload AS jsonb),
                                :createdAt)
                        """)
                .param("eventId", envelope.eventId())
                .param("topic", type.topic())
                .param("messageKey", messageKey)
                .param("aggregateId", rideId != null ? rideId : driverId)
                .param("eventType", envelope.eventType())
                .param("payload", codec.encode(envelope))
                .param("createdAt", Timestamps.toDb(envelope.occurredAt()))
                .update();
        try (LogContext ignored = LogContext.with(LogContext.EVENT_ID, envelope.eventId())
                .and(LogContext.EVENT_TYPE, envelope.eventType())) {
            log.info("Event {} recorded in outbox for {}", envelope.eventType(), type.topic());
        }
        return envelope.eventId();
    }
}
