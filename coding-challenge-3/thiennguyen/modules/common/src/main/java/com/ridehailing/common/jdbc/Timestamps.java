package com.ridehailing.common.jdbc;

import java.sql.ResultSet;
import java.sql.SQLException;
import java.time.Instant;
import java.time.OffsetDateTime;
import java.time.ZoneOffset;

/**
 * The PostgreSQL driver binds {@link OffsetDateTime} for TIMESTAMPTZ, not {@link Instant}.
 */
public final class Timestamps {

    private Timestamps() {
    }

    public static OffsetDateTime toDb(Instant instant) {
        return OffsetDateTime.ofInstant(instant, ZoneOffset.UTC);
    }

    public static Instant fromDb(ResultSet resultSet, String column) throws SQLException {
        OffsetDateTime value = resultSet.getObject(column, OffsetDateTime.class);
        return value == null ? null : value.toInstant();
    }
}
