package com.ridehailing.dispatch.application.port;

import com.ridehailing.dispatch.domain.OfferTransition;
import com.ridehailing.dispatch.domain.RideOffer;
import java.time.Instant;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

/** Offers of a ride to drivers; at most one is PENDING per ride and per driver. */
public interface OfferRepository {

    void insert(RideOffer offer);

    /** Accepts the driver's pending offer, only while it has not expired at {@code now}. */
    boolean accept(UUID rideId, String driverId, Instant now);

    /** Resolves the driver's pending offer with {@code transition} (rejected, cancelled, or expired early). */
    boolean resolvePending(UUID rideId, String driverId, OfferTransition transition, Instant now);

    /** Expires the offer if it is still pending and its response timeout has passed at {@code now}. */
    boolean expireIfDue(UUID offerId, Instant now);

    List<RideOffer> findDueForExpiry(Instant now, int limit);

    Optional<RideOffer> findPending(UUID rideId);

    /** The most recent offer of this ride to this driver, whatever its outcome. */
    Optional<RideOffer> findLatest(UUID rideId, String driverId);

    List<RideOffer> findByRide(UUID rideId);
}
