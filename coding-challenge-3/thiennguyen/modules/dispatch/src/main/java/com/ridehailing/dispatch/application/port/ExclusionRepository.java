package com.ridehailing.dispatch.application.port;

import com.ridehailing.dispatch.domain.OfferStatus;
import java.time.Instant;
import java.util.Set;
import java.util.UUID;

/** Drivers who rejected or let an offer expire are never offered that ride again. */
public interface ExclusionRepository {

    void add(UUID rideId, String driverId, OfferStatus reason, Instant now);

    Set<String> forRide(UUID rideId);

    boolean contains(UUID rideId, String driverId);
}
