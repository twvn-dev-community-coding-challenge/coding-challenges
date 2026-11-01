package com.ridehailing.driver.application;

import com.ridehailing.common.error.ConflictException;
import com.ridehailing.common.error.NotFoundException;
import com.ridehailing.common.error.ValidationException;
import com.ridehailing.common.events.OutboxWriter;
import com.ridehailing.common.kernel.geo.Location;
import com.ridehailing.common.logging.LogContext;
import com.ridehailing.contracts.DriverAvailabilityChangedV1;
import com.ridehailing.contracts.DriverLocationUpdatedV1;
import com.ridehailing.contracts.EventType;
import com.ridehailing.contracts.GeoPoint;
import com.ridehailing.driver.application.port.DriverReadCache;
import com.ridehailing.driver.application.port.DriverRepository;
import com.ridehailing.driver.domain.Driver;
import com.ridehailing.driver.domain.DriverStatus;
import com.ridehailing.driver.domain.DriverTransition;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.Map;

/**
 * Driver-owned facts: availability and location. Each change is announced on driver-events so dispatch keeps its supply
 * copy current.
 */
@Service
@RequiredArgsConstructor
@Slf4j
public class DriverService {

    private final DriverRepository drivers;
    private final DriverReadCache cache;
    private final OutboxWriter outbox;

    public Driver get(String driverId) {
        return cache.get(driverId).orElseGet(() -> {
            Driver driver = drivers.find(driverId).orElseThrow(() -> notFound(driverId));
            cache.put(driver);
            return driver;
        });
    }

    /**
     * Decided on the driver's own copy of its assignment. If dispatch has just made an offer this copy has not seen
     * yet, going offline still succeeds, and dispatch treats the offline event as the offer expiring.
     */
    @Transactional
    public Driver changeAvailability(String driverId, DriverStatus requested) {
        DriverTransition transition = DriverTransition.toward(requested).orElseThrow(
                () -> new ValidationException(
                        "UNSUPPORTED_AVAILABILITY", "availability can only be set to AVAILABLE or OFFLINE"));
        Driver current = drivers.findForUpdate(driverId).orElseThrow(() -> notFound(driverId));
        try (LogContext ignored = LogContext.with(LogContext.DRIVER_ID, driverId)) {
            return switch (transition.evaluate(current.status())) {
                case APPLY -> {
                    Driver updated = drivers.applyAvailability(driverId, transition)
                            .orElseThrow(() -> new IllegalStateException("Locked driver changed concurrently"));
                    cache.putAfterCommit(updated);
                    outbox.appendForDriver(EventType.DRIVER_AVAILABILITY_CHANGED, driverId, null,
                            new DriverAvailabilityChangedV1(
                                    driverId, updated.availability() == DriverStatus.AVAILABLE, updated.version()));
                    log.info("Driver availability {} -> {}", current.status(), transition.to());
                    yield updated;
                }
                case ALREADY_APPLIED -> current;
                case REJECTED -> throw new ConflictException(
                        "INVALID_DRIVER_TRANSITION",
                        "Driver is %s and cannot become %s".formatted(current.status(), requested),
                        Map.of("currentStatus", current.status().name()));
            };
        }
    }

    @Transactional
    public Driver updateLocation(String driverId, Location location) {
        Driver updated = drivers.updateLocation(driverId, location).orElseThrow(() -> notFound(driverId));
        cache.putAfterCommit(updated);
        outbox.appendForDriver(EventType.DRIVER_LOCATION_UPDATED, driverId, null, new DriverLocationUpdatedV1(
                driverId, new GeoPoint(location.latitude(), location.longitude()), updated.version()));
        try (LogContext ignored = LogContext.with(LogContext.DRIVER_ID, driverId)) {
            log.info("Driver location updated");
        }
        return updated;
    }

    private static NotFoundException notFound(String driverId) {
        return new NotFoundException("DRIVER_NOT_FOUND", "Driver %s does not exist".formatted(driverId));
    }
}
