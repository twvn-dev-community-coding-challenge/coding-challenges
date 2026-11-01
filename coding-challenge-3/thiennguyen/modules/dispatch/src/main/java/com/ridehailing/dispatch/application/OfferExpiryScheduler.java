package com.ridehailing.dispatch.application;

import com.ridehailing.contracts.OfferExpiredV1;
import com.ridehailing.dispatch.application.port.OfferRepository;
import com.ridehailing.dispatch.domain.RideOffer;
import java.time.Clock;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;

/** Expires pending offers past their response timeout; each expiry is its own transaction. */
@Component
@RequiredArgsConstructor
@Slf4j
public class OfferExpiryScheduler {

    private static final int BATCH_SIZE = 100;

    private final OfferRepository offers;
    private final OfferService offerService;
    private final Clock clock;

    @Scheduled(fixedDelayString = "${dispatch.expiry-poll-interval-ms:1000}")
    public void expireDueOffers() {
        for (RideOffer offer : offers.findDueForExpiry(clock.instant(), BATCH_SIZE)) {
            try {
                offerService.expire(offer, OfferExpiredV1.TIMEOUT);
            } catch (RuntimeException exception) {
                log.warn("Could not expire offer {}: {}", offer.id(), exception.toString());
            }
        }
    }
}
