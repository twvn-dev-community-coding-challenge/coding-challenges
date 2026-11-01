package com.ridehailing.dispatch.domain;

import static org.assertj.core.api.Assertions.assertThat;

import java.time.Duration;
import java.time.Instant;
import java.util.UUID;
import org.junit.jupiter.api.Test;

class RideOfferTest {

    @Test
    void aPendingOfferExpiresAfterTheTimeoutAndHasNoAnswerYet() {
        UUID rideId = UUID.randomUUID();
        Instant now = Instant.parse("2026-10-26T10:00:00Z");

        RideOffer offer = RideOffer.pending(rideId, "D001", 0.8, now, Duration.ofSeconds(30));

        assertThat(offer.id()).isNotNull();
        assertThat(offer.rideId()).isEqualTo(rideId);
        assertThat(offer.driverId()).isEqualTo("D001");
        assertThat(offer.status()).isEqualTo(OfferStatus.PENDING);
        assertThat(offer.distanceKm()).isEqualTo(0.8);
        assertThat(offer.createdAt()).isEqualTo(now);
        assertThat(offer.expiresAt()).isEqualTo(now.plusSeconds(30));
        assertThat(offer.respondedAt()).isNull();
    }

    @Test
    void everyPendingOfferHasItsOwnId() {
        Instant now = Instant.parse("2026-10-26T10:00:00Z");
        UUID rideId = UUID.randomUUID();

        assertThat(RideOffer.pending(rideId, "D001", 0.8, now, Duration.ofSeconds(30)).id())
                .isNotEqualTo(RideOffer.pending(rideId, "D001", 0.8, now, Duration.ofSeconds(30)).id());
    }
}
