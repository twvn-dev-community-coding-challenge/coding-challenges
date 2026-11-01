package com.ridehailing.driver.adapters.cache;

import com.ridehailing.common.cache.CacheProperties;
import com.ridehailing.common.cache.VersionedCache;
import com.ridehailing.common.cache.VersionedJsonCache;
import com.ridehailing.driver.application.port.DriverReadCache;
import com.ridehailing.driver.domain.Driver;
import org.springframework.stereotype.Component;
import tools.jackson.databind.ObjectMapper;

@Component
public class RedisDriverReadCache extends VersionedJsonCache<String, Driver> implements DriverReadCache {

    public RedisDriverReadCache(VersionedCache cache, ObjectMapper objectMapper, CacheProperties properties) {
        super(cache, objectMapper, "driver", Driver.class, properties.driverTtl(), Driver::id, Driver::version);
    }
}
