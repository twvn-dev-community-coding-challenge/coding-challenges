package com.ridehailing.ride.application.port;

/**
 * What ride needs to know about passengers: whether one exists. Passengers are seeded; registration is out of scope.
 */
public interface Passengers {

    boolean exists(String passengerId);
}
