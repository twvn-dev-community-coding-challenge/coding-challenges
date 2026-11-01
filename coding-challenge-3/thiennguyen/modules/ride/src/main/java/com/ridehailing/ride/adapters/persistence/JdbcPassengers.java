package com.ridehailing.ride.adapters.persistence;

import com.ridehailing.ride.application.port.Passengers;
import lombok.RequiredArgsConstructor;
import org.springframework.jdbc.core.simple.JdbcClient;
import org.springframework.stereotype.Repository;

/**
 * Passengers are seeded records; registration is out of scope.
 */
@Repository
@RequiredArgsConstructor
public class JdbcPassengers implements Passengers {

    private final JdbcClient jdbc;

    @Override
    public boolean exists(String passengerId) {
        return jdbc.sql("SELECT EXISTS (SELECT 1 FROM passengers WHERE id = :id)")
                .param("id", passengerId)
                .query(Boolean.class)
                .single();
    }
}
