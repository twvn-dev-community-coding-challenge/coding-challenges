package com.ridehailing.contracts;

/**
 * A WGS-84 coordinate on the wire. Location is personal data: keep it out of logs (GDPR).
 */
public record GeoPoint(double latitude, double longitude) {
}
