package com.ridehailing.common.events;

import com.ridehailing.contracts.EventEnvelope;
import java.util.Map;

/**
 * Converts events between their wire form and records. Decoding failures are {@link MalformedEventException}s, which
 * are dead-lettered without retries.
 */
public interface EventCodec {

    String encode(EventEnvelope envelope);

    /** The JSON object form of a {@code contracts} payload record, as carried in {@link EventEnvelope#payload()}. */
    Map<String, Object> toPayload(Object payloadRecord);

    EventEnvelope decode(String json);

    /** The envelope's payload as the record {@code payloadType}. */
    <T> T payload(EventEnvelope envelope, Class<T> payloadType);
}
