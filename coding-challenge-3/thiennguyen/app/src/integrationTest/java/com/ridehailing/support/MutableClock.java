package com.ridehailing.support;

import java.time.Clock;
import java.time.Duration;
import java.time.Instant;
import java.time.ZoneId;
import java.time.ZoneOffset;
import java.util.concurrent.atomic.AtomicReference;

/** Real time plus an offset tests can advance, e.g. past an offer's response timeout. */
public class MutableClock extends Clock {

    private final AtomicReference<Duration> offset = new AtomicReference<>(Duration.ZERO);

    @Override
    public ZoneId getZone() {
        return ZoneOffset.UTC;
    }

    @Override
    public Clock withZone(ZoneId zone) {
        return this;
    }

    @Override
    public Instant instant() {
        return Instant.now().plus(offset.get());
    }

    public void advance(Duration duration) {
        offset.updateAndGet(current -> current.plus(duration));
    }
}
