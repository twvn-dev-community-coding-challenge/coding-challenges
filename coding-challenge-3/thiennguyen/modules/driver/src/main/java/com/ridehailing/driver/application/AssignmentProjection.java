package com.ridehailing.driver.application;

import com.ridehailing.common.events.EventCodec;
import com.ridehailing.common.events.ProcessedEvents;
import com.ridehailing.contracts.DriverAssignmentChangedV1;
import com.ridehailing.contracts.EventEnvelope;
import com.ridehailing.contracts.EventType;
import com.ridehailing.driver.application.port.DriverReadCache;
import com.ridehailing.driver.application.port.DriverRepository;
import com.ridehailing.driver.domain.Assignment;
import com.ridehailing.driver.domain.DriverAssignment;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;

import java.util.Set;

/**
 * Keeps the driver's copy of dispatch's assignment (OFFERED/BUSY/none); versioned, so order and duplicates are safe.
 */
@Component
@RequiredArgsConstructor
@Slf4j
public class AssignmentProjection {

    public static final String CONSUMER = "driver";
    public static final Set<EventType> HANDLED_EVENTS = Set.of(EventType.DRIVER_ASSIGNMENT_CHANGED);

    private final DriverRepository drivers;
    private final DriverReadCache cache;
    private final ProcessedEvents processedEvents;
    private final EventCodec codec;

    @Transactional
    public void apply(EventEnvelope envelope) {
        if (!processedEvents.markProcessed(CONSUMER, envelope.eventId())) {
            log.info("Duplicate delivery ignored; assignment was already applied");
            return;
        }
        DriverAssignmentChangedV1 change = codec.payload(envelope, DriverAssignmentChangedV1.class);
        DriverAssignment assignment = DriverAssignment.of(Assignment.valueOf(change.assignment().name()), change.rideId());
        drivers.applyAssignment(change.driverId(), assignment, change.version()).ifPresentOrElse(
                driver -> {
                    cache.putAfterCommit(driver);
                    log.info("Driver assignment is now {} (dispatch version {})", change.assignment(), change.version());
                },
                () -> log.info("Ignored assignment {} at dispatch version {}: newer one already applied",
                        change.assignment(), change.version()));
    }
}
