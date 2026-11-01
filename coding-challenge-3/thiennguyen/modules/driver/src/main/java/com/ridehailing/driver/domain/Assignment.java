package com.ridehailing.driver.domain;

/**
 * Whether dispatch holds the driver: not at all, for a pending offer, or for a trip (dispatch decides).
 */
public enum Assignment {
    NONE,
    OFFERED,
    BUSY
}
