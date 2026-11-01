package com.ridehailing.ride.application;

import com.ridehailing.common.kernel.geo.Location;

public record CreateRideCommand(String passengerId, Location pickup, Location destination) {
}
