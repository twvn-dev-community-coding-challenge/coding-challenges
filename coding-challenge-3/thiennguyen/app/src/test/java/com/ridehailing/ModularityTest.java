package com.ridehailing;

import org.junit.jupiter.api.Test;
import org.springframework.modulith.core.ApplicationModules;
import org.springframework.modulith.core.VerificationOptions;
import org.springframework.modulith.docs.Documenter;

/**
 * The architecture check: Spring Modulith verifies module boundaries (no cycles, no access to another module's
 * sub-packages, only the {@code allowedDependencies} each module declares) together with the layering rules of
 * {@link ArchitectureRules}. A second test writes the module documentation generated from the same model.
 */
class ModularityTest {

    static final ApplicationModules MODULES = ApplicationModules.of(RideHailingApplication.class);

    @Test
    void modulesRespectTheirBoundariesAndLayering() {
        MODULES.verify(VerificationOptions.defaults().andAdditionalVerifications(ArchitectureRules.all()));
    }

    /**
     * Writes PlantUML component diagrams (whole system and one per module) and a canvas per module to
     * {@code app/build/spring-modulith-docs}. Ride, driver and dispatch appear unconnected: they talk through Kafka,
     * which Modulith cannot see; the event flow is in docs/architecture.md § Event catalogue.
     */
    @Test
    void writesModuleDocumentation() {
        new Documenter(MODULES, Documenter.Options.defaults().withOutputFolder("build/spring-modulith-docs"))
                .writeDocumentation()
                .writeIndividualModulesAsPlantUml();
    }
}
