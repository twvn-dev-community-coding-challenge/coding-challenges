package com.ridehailing.dispatch.application;

import com.ridehailing.common.kernel.geo.Location;
import com.ridehailing.common.logging.LogContext;
import com.ridehailing.contracts.DriverAvailabilityChangedV1;
import com.ridehailing.contracts.DriverLocationUpdatedV1;
import com.ridehailing.contracts.OfferExpiredV1;
import com.ridehailing.dispatch.application.port.DriverSupplyRepository;
import com.ridehailing.dispatch.application.port.OfferRepository;
import com.ridehailing.dispatch.domain.Assignment;
import com.ridehailing.dispatch.domain.SupplyDriver;
import java.util.Optional;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Propagation;
import org.springframework.transaction.annotation.Transactional;

/**
 * Applies driver-events to dispatch's supply copy. A driver who goes offline while holding a pending offer — possible
 * because the driver module checks its own, slightly older copy of the assignment — has that offer expired and the ride
 * re-matched. A driver who goes offline mid-trip keeps the trip and just stops receiving offers.
 */
@Component
@RequiredArgsConstructor
@Slf4j
public class SupplyUpdates {

    private final DriverSupplyRepository supply;
    private final OfferRepository offers;
    private final OfferService offerService;

    @Transactional(propagation = Propagation.MANDATORY)
    public void availabilityChanged(DriverAvailabilityChangedV1 change) {
        try (LogContext ignored = LogContext.with(LogContext.DRIVER_ID, change.driverId())) {
            Optional<SupplyDriver> current = supply.find(change.driverId());
            boolean stale = current.isPresent() && current.get().availabilityVersion() >= change.version();
            if (stale) {
                log.info("Ignored availability at driver version {}: newer one already applied", change.version());
                return;
            }
            if (!change.online() && current.isPresent() && current.get().assignment() == Assignment.OFFERED) {
                offers.findPending(current.get().currentRideId()).ifPresent(
                        offer -> offerService.expire(offer, OfferExpiredV1.DRIVER_OFFLINE));
            }
            supply.applyAvailability(change.driverId(), change.online(), change.version());
            log.info("Driver is now {} for dispatch", change.online() ? "online" : "offline");
        }
    }

    @Transactional(propagation = Propagation.MANDATORY)
    public void locationUpdated(DriverLocationUpdatedV1 change) {
        Location location = new Location(change.location().latitude(), change.location().longitude());
        supply.applyLocation(change.driverId(), location, change.version());
    }
}
