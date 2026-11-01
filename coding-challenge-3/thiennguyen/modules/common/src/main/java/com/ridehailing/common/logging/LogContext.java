package com.ridehailing.common.logging;

import java.util.LinkedHashMap;
import java.util.Map;
import java.util.UUID;
import org.slf4j.MDC;

/**
 * Puts correlation identifiers into the logging MDC for the duration of a try-with-resources block, restoring the
 * previous values on close. The console pattern prints rideId, driverId and eventId on every line
 * ({@code CONSOLE_LOG_PATTERN} in {@code logback-spring.xml}), so a ride's lines can be found with grep. Never put
 * coordinates or other personal data in here.
 */
public final class LogContext implements AutoCloseable {

    public static final String RIDE_ID = "rideId";
    public static final String DRIVER_ID = "driverId";
    public static final String EVENT_ID = "eventId";
    public static final String EVENT_TYPE = "eventType";

    private final Map<String, String> previousValues = new LinkedHashMap<>();

    private LogContext() {
    }

    public static LogContext with(String key, Object value) {
        return new LogContext().and(key, value);
    }

    public static LogContext ride(UUID rideId) {
        return with(RIDE_ID, rideId);
    }

    public static LogContext rideAndDriver(UUID rideId, String driverId) {
        return with(RIDE_ID, rideId).and(DRIVER_ID, driverId);
    }

    public LogContext and(String key, Object value) {
        if (value == null) {
            return this;
        }
        if (!previousValues.containsKey(key)) {
            previousValues.put(key, MDC.get(key));
        }
        MDC.put(key, value.toString());
        return this;
    }

    @Override
    public void close() {
        previousValues.forEach((key, previous) -> {
            if (previous == null) {
                MDC.remove(key);
            } else {
                MDC.put(key, previous);
            }
        });
    }
}
