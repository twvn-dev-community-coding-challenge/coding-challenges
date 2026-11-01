package com.ridehailing.dispatch.domain;

/** Whether dispatch holds a driver: free, for a pending offer, or for a trip. Dispatch alone decides it. */
public enum Assignment {
    NONE,
    OFFERED,
    BUSY
}
