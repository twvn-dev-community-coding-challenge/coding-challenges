package com.ridehailing.common.events;

import com.ridehailing.common.jdbc.Timestamps;
import java.time.Clock;
import java.util.UUID;
import lombok.RequiredArgsConstructor;
import org.springframework.jdbc.core.simple.JdbcClient;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Propagation;
import org.springframework.transaction.annotation.Transactional;

/** {@link ProcessedEvents} on the {@code processed_events} table: the insert is the check. */
@Component
@RequiredArgsConstructor
public class JdbcProcessedEvents implements ProcessedEvents {

    private final JdbcClient jdbc;
    private final Clock clock;

    @Override
    @Transactional(propagation = Propagation.MANDATORY)
    public boolean markProcessed(String consumer, UUID eventId) {
        return jdbc.sql("""
                        INSERT INTO processed_events (consumer, event_id, processed_at)
                        VALUES (:consumer, :eventId, :processedAt)
                        ON CONFLICT DO NOTHING
                        """)
                .param("consumer", consumer)
                .param("eventId", eventId)
                .param("processedAt", Timestamps.toDb(clock.instant()))
                .update() == 1;
    }
}
