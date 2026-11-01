/**
 * Dispatch (core domain): dispatch requests, matching, offers, exclusions and the driver assignment. Exposes no API to
 * other modules: it integrates only through events in {@code contracts}, so it may depend on the shared {@code common}
 * module and the published language alone, and can be extracted as a service.
 */
@ApplicationModule(displayName = "Dispatch", allowedDependencies = {"common", "contracts"})
package com.ridehailing.dispatch;

import org.springframework.modulith.ApplicationModule;
