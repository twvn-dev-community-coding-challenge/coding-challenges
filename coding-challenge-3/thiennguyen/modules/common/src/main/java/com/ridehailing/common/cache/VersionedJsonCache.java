package com.ridehailing.common.cache;

import java.time.Duration;
import java.util.Optional;
import java.util.function.Function;
import java.util.function.ToLongFunction;
import tools.jackson.core.JacksonException;
import tools.jackson.databind.ObjectMapper;

/**
 * A read model cached as JSON under {@code rh:<name>:<id>} through {@link VersionedCache}: a value is stored only if its
 * version is newer than the cached one, and an entry that no longer decodes counts as a miss.
 */
public abstract class VersionedJsonCache<K, T> {

    private final VersionedCache cache;
    private final ObjectMapper objectMapper;
    private final String name;
    private final Class<T> type;
    private final Duration ttl;
    private final Function<T, K> idOf;
    private final ToLongFunction<T> versionOf;

    protected VersionedJsonCache(VersionedCache cache, ObjectMapper objectMapper, String name, Class<T> type,
            Duration ttl, Function<T, K> idOf, ToLongFunction<T> versionOf) {
        this.cache = cache;
        this.objectMapper = objectMapper;
        this.name = name;
        this.type = type;
        this.ttl = ttl;
        this.idOf = idOf;
        this.versionOf = versionOf;
    }

    public static String key(String name, Object id) {
        return "rh:" + name + ":" + id;
    }

    public Optional<T> get(K id) {
        return cache.get(name, key(name, id)).flatMap(this::decode);
    }

    public void put(T value) {
        cache.putIfNewer(name, key(name, idOf.apply(value)), versionOf.applyAsLong(value),
                objectMapper.writeValueAsString(value), ttl);
    }

    public void putAfterCommit(T value) {
        cache.putIfNewerAfterCommit(name, key(name, idOf.apply(value)), versionOf.applyAsLong(value),
                objectMapper.writeValueAsString(value), ttl);
    }

    private Optional<T> decode(String json) {
        try {
            return Optional.of(objectMapper.readValue(json, type));
        } catch (JacksonException exception) {
            return Optional.empty();
        }
    }
}
