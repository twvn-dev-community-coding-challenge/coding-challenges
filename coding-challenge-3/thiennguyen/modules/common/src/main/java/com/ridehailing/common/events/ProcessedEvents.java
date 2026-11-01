package com.ridehailing.common.events;

import java.util.UUID;

/**
 * Consumer-side de-duplication, inside the consumer's own transaction: the marker commits with the consumer's work, or
 * neither does. Each call must run in a transaction.
 */
public interface ProcessedEvents {

    /** Records that {@code consumer} processed the event; returns {@code false} when it already had. */
    boolean markProcessed(String consumer, UUID eventId);
}
