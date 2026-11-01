package com.ridehailing.common.events;

import com.ridehailing.contracts.EventEnvelope;
import java.util.Map;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Component;
import tools.jackson.core.JacksonException;
import tools.jackson.databind.ObjectMapper;

/** {@link EventCodec} with Jackson, the application's own {@link ObjectMapper}. */
@Component
@RequiredArgsConstructor
public class JacksonEventCodec implements EventCodec {

    private final ObjectMapper objectMapper;

    @Override
    public String encode(EventEnvelope envelope) {
        return objectMapper.writeValueAsString(envelope);
    }

    @Override
    @SuppressWarnings("unchecked")
    public Map<String, Object> toPayload(Object payloadRecord) {
        return objectMapper.convertValue(payloadRecord, Map.class);
    }

    @Override
    public EventEnvelope decode(String json) {
        EventEnvelope envelope;
        try {
            envelope = objectMapper.readValue(json, EventEnvelope.class);
        } catch (JacksonException exception) {
            throw new MalformedEventException("Event is not a valid envelope", exception);
        }
        if (envelope == null || envelope.eventId() == null || envelope.eventType() == null
                || envelope.payload() == null) {
            throw new MalformedEventException("Event envelope is missing eventId, eventType or payload");
        }
        return envelope;
    }

    @Override
    public <T> T payload(EventEnvelope envelope, Class<T> payloadType) {
        try {
            return objectMapper.convertValue(envelope.payload(), payloadType);
        } catch (IllegalArgumentException exception) {
            throw new MalformedEventException(
                    "Payload of %s is not a valid %s".formatted(envelope.eventType(), payloadType.getSimpleName()),
                    exception);
        }
    }
}
