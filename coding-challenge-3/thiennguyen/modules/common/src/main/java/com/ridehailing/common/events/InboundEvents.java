package com.ridehailing.common.events;

import com.ridehailing.common.logging.LogContext;
import com.ridehailing.contracts.EventEnvelope;
import com.ridehailing.contracts.EventType;
import io.micrometer.core.instrument.MeterRegistry;

import java.time.Clock;
import java.time.Duration;
import java.util.Optional;
import java.util.Set;
import java.util.function.BiConsumer;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.apache.kafka.clients.consumer.ConsumerRecord;
import org.springframework.stereotype.Component;

/**
 * What every Kafka listener does before its handler runs: decode the envelope, skip the event types this consumer
 * does not handle, and keep the event's identifiers in the logging context while it is handled. A record that cannot be
 * decoded is a {@link MalformedEventException}, which is dead-lettered without retries.
 * <p>
 * {@code ridehailing.events.latency{consumer, event}} times each handled event from being recorded in the outbox to its
 * consumer starting on it: outbox wait, relay, broker and consumer queue together.
 */
@Component
@RequiredArgsConstructor
@Slf4j
public class InboundEvents {
    private final EventCodec codec;
    private final Clock clock;
    private final MeterRegistry meters;

    public void handle(String consumer, ConsumerRecord<String, String> record, Set<EventType> handledEvents,
                       BiConsumer<EventType, EventEnvelope> handler) {
        if (record.value() == null) {
            throw new MalformedEventException("Empty record on " + record.topic());
        }
        EventEnvelope envelope = codec.decode(record.value());
        Optional<EventType> type = EventType.fromWireName(envelope.eventType()).filter(handledEvents::contains);
        if (type.isEmpty()) {
            return;
        }
        meters.timer("ridehailing.events.latency", "consumer", consumer, "event", envelope.eventType())
                .record(Duration.between(envelope.occurredAt(), clock.instant()));
        try (LogContext ignored = LogContext.with(LogContext.EVENT_ID, envelope.eventId())
                .and(LogContext.RIDE_ID, envelope.rideId())
                .and(LogContext.DRIVER_ID, envelope.driverId())
                .and(LogContext.EVENT_TYPE, envelope.eventType())) {
            log.info("{} consuming {} ({} partition {}, offset {})",
                    consumer, envelope.eventType(), record.topic(), record.partition(), record.offset());
            handler.accept(type.get(), envelope);
        }
    }
}
