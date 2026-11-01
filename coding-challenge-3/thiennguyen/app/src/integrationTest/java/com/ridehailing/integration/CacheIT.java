package com.ridehailing.integration;

import static org.assertj.core.api.Assertions.assertThat;

import com.ridehailing.common.cache.VersionedCache;
import com.ridehailing.driver.domain.DriverStatus;
import com.ridehailing.ride.adapters.cache.RedisRideReadCache;
import com.ridehailing.ride.adapters.web.model.RideResponse;
import com.ridehailing.ride.domain.RideStatus;
import com.ridehailing.support.Api;
import com.ridehailing.support.IntegrationTestSupport;
import com.ridehailing.support.TestCity;
import java.time.Duration;
import java.util.UUID;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.testcontainers.containers.GenericContainer;

/** Redis only answers reads, never holds a stale newer-than-truth value, and is optional for correctness. */
@Tag("cache")
class CacheIT extends IntegrationTestSupport {

    @Autowired
    private StringRedisTemplate redis;

    @Autowired
    private VersionedCache cache;

    @Autowired
    @Qualifier("redis")
    private GenericContainer<?> redisContainer;

    private final TestCity city = TestCity.next();

    @Test
    void readingARideCachesItWithItsVersion() {
        UUID rideId = createRide(fixtures.passenger(), city.centre()).id();

        api.get("/rides/" + rideId);

        eventually().untilAsserted(() -> assertThat(cachedVersion(rideId)).isEqualTo(rideRow(rideId).version()));
    }

    @Test
    void everyCommittedStateChangeRefreshesTheCachedRide() {
        String driver = fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(0.5));
        UUID rideId = createRide(fixtures.passenger(), city.centre()).id();
        awaitOfferedTo(rideId, driver);

        acceptAndAwait(rideId, driver);

        eventually().untilAsserted(() -> assertThat(cachedVersion(rideId)).isEqualTo(rideRow(rideId).version()));
        assertThat(api.get("/rides/" + rideId).as(RideResponse.class).status()).isEqualTo(RideStatus.DRIVER_ACCEPTED);
    }

    @Test
    void anOlderVersionNeverOverwritesANewerOne() {
        UUID rideId = createRide(fixtures.passenger(), city.centre()).id();
        api.get("/rides/" + rideId);
        eventually().untilAsserted(() -> assertThat(cachedVersion(rideId)).isNotNull());

        boolean stored = cache.putIfNewer("ride", RedisRideReadCache.key(rideId), -1, "{\"stale\":true}", Duration.ofMinutes(1));

        assertThat(stored).isFalse();
        assertThat(api.get("/rides/" + rideId).as(RideResponse.class).id()).isEqualTo(rideId);
    }

    @Test
    void everyFlowStillWorksWhileRedisIsUnavailable() {
        String driver = fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(0.5));
        redisContainer.getDockerClient().pauseContainerCmd(redisContainer.getContainerId()).exec();
        try {
            UUID rideId = createRide(fixtures.passenger(), city.centre()).id();
            awaitOfferedTo(rideId, driver);

            Api.Response accepted = accept(rideId, driver);
            assertThat(accepted.status()).as(accepted.body()).isEqualTo(200);

            eventually().untilAsserted(() -> {
                Api.Response read = api.get("/rides/" + rideId);
                assertThat(read.status()).as(read.body()).isEqualTo(200);
                assertThat(read.as(RideResponse.class).status()).isEqualTo(RideStatus.DRIVER_ACCEPTED);
            });
        } finally {
            redisContainer.getDockerClient().unpauseContainerCmd(redisContainer.getContainerId()).exec();
        }
        assertInvariants();
    }

    private Long cachedVersion(UUID rideId) {
        String version = redis.<String, String>opsForHash().get(RedisRideReadCache.key(rideId), "v");
        return version == null ? null : Long.valueOf(version);
    }
}
