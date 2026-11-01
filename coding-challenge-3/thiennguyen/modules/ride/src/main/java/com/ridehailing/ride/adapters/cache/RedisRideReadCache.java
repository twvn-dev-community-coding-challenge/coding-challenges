package com.ridehailing.ride.adapters.cache;

import com.ridehailing.common.cache.CacheProperties;
import com.ridehailing.common.cache.VersionedCache;
import com.ridehailing.common.cache.VersionedJsonCache;
import com.ridehailing.ride.application.port.RideReadCache;
import com.ridehailing.ride.domain.Ride;
import org.springframework.stereotype.Component;
import tools.jackson.databind.ObjectMapper;

import java.util.UUID;

@Component
public class RedisRideReadCache extends VersionedJsonCache<UUID, Ride> implements RideReadCache {

    private static final String NAME = "ride";

    public RedisRideReadCache(VersionedCache cache, ObjectMapper objectMapper, CacheProperties properties) {
        super(cache, objectMapper, NAME, Ride.class, properties.rideTtl(), Ride::id, Ride::version);
    }

    public static String key(UUID rideId) {
        return key(NAME, rideId);
    }
}
