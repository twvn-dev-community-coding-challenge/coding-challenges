/**
 * Driver: availability and location, plus a versioned copy of dispatch's assignment. Exposes no API to other modules:
 * dispatch learns about drivers only from events in {@code contracts}.
 */
@ApplicationModule(displayName = "Driver", allowedDependencies = {"common", "contracts"})
package com.ridehailing.driver;

import org.springframework.modulith.ApplicationModule;
