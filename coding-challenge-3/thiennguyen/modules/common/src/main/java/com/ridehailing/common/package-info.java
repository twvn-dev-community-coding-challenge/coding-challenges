/**
 * Shared building blocks used by every module: the shared kernel ({@code common.kernel}: geo, lifecycle transitions)
 * and platform code (API errors, the transactional outbox and de-duplication, the Redis cache, logging context).
 * Declared open so other modules may use its sub-packages directly.
 */
@ApplicationModule(displayName = "Common", type = ApplicationModule.Type.OPEN, allowedDependencies = "contracts")
package com.ridehailing.common;

import org.springframework.modulith.ApplicationModule;
