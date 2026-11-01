package com.ridehailing.contracts;

/**
 * Whether dispatch has a driver tied to a ride: free, holding a pending offer, or on an accepted trip.
 */
public enum AssignmentState {
    NONE,
    OFFERED,
    BUSY
}
