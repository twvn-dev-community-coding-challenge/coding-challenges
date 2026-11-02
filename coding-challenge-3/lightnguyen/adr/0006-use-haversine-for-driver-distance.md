# ADR-0006: Use Haversine for driver distance

- Status: Accepted
- Date: 2026-10-31

## Context

Driver matching needs to compare driver and ride-pickup locations represented by latitude and longitude, exclude drivers outside a configured radius, and order eligible drivers by distance. The challenge requires only a simple geographic distance calculation.

## Decision

Use the Haversine formula to calculate great-circle distance between coordinates on Earth. Use an Earth radius of 6,371 km and compare the resulting distance in kilometers with the configured matching radius.

## Consequences

- The calculation accounts for Earth's curvature and is straightforward to implement without a geospatial database or additional dependency.
- It is sufficiently accurate for the challenge's nearby-driver matching use case.
- It models Earth as a sphere, so it is an approximation rather than a survey-grade ellipsoidal geodesic. A higher-precision geospatial solution may be needed for applications with stricter accuracy requirements.
