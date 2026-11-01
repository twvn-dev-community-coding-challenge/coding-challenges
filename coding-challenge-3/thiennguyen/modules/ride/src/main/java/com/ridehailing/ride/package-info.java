/**
 * Ride: passengers (seeded), ride creation (idempotent), lifecycle, history and the read API. Exposes no API to other
 * modules: dispatch and driver learn about rides only from events in {@code contracts}.
 */
@ApplicationModule(displayName = "Ride", allowedDependencies = {"common", "contracts"})
package com.ridehailing.ride;

import org.springframework.modulith.ApplicationModule;
