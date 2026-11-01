package com.ridehailing;

import com.tngtech.archunit.lang.ArchRule;

import java.util.List;

import static com.tngtech.archunit.base.DescribedPredicate.not;
import static com.tngtech.archunit.core.domain.JavaClass.Predicates.INTERFACES;
import static com.tngtech.archunit.core.domain.JavaClass.Predicates.resideInAPackage;
import static com.tngtech.archunit.lang.syntax.ArchRuleDefinition.*;
import static com.tngtech.archunit.library.Architectures.layeredArchitecture;
import static com.tngtech.archunit.library.dependencies.SlicesRuleDefinition.slices;

/**
 * The rules Spring Modulith cannot express on its own, run by {@link ModularityTest} as part of
 * {@code ApplicationModules.verify()}. Which modules may depend on which is declared in each module's
 * {@code package-info.java} ({@code @ApplicationModule(allowedDependencies = …)}) and checked by Modulith.
 */
final class ArchitectureRules {

    /**
     * Clean Architecture's dependency rule inside ride, driver and dispatch: source code depends only inwards, from
     * the frameworks to the adapters, from the adapters to the application, and from the application to the domain.
     */
    static final ArchRule DEPENDENCIES_POINT_INWARDS = layeredArchitecture().consideringOnlyDependenciesInLayers()
            .layer("Domain").definedBy("com.ridehailing.*.domain..")
            .layer("Application").definedBy("com.ridehailing.*.application..")
            .layer("Adapters").definedBy("com.ridehailing.*.adapters..")
            .layer("Frameworks").definedBy("com.ridehailing.*.frameworks..")
            .whereLayer("Frameworks").mayNotBeAccessedByAnyLayer()
            .whereLayer("Adapters").mayOnlyBeAccessedByLayers("Frameworks")
            .whereLayer("Application").mayOnlyBeAccessedByLayers("Adapters", "Frameworks")
            .as("Dependencies point inwards: frameworks → adapters → application → domain");

    /**
     * Web, messaging, persistence and cache adapters meet only in the application, never in each other.
     */
    static final ArchRule ADAPTERS_DO_NOT_DEPEND_ON_EACH_OTHER = slices()
            .matching("com.ridehailing.(*).adapters.(*)..")
            .should().notDependOnEachOther()
            .as("Each kind of adapter talks only to the application, not to another kind of adapter");

    /**
     * The ports are what the application needs from the outside world; adapters implement them.
     */
    static final ArchRule PORTS_ARE_INTERFACES = classes()
            .that().resideInAPackage("com.ridehailing.*.application.port..").and().areTopLevelClasses()
            .should().beInterfaces()
            .as("Application ports are interfaces");

    /**
     * The domain uses only {@code java..}, the domain packages and the shared kernel: no Spring, SQL, Kafka, Redis,
     * HTTP, Jackson, Lombok or {@code contracts}. Modulith already keeps one context out of another's domain.
     */
    static final ArchRule DOMAIN_IS_PLAIN_JAVA = classes()
            .that().resideInAPackage("com.ridehailing.*.domain..")
            .should().onlyDependOnClassesThat().resideInAnyPackage(
                    "java..", "com.ridehailing.*.domain..", "com.ridehailing.common.kernel..")
            .as("The domain is plain Java");

    static final ArchRule SHARED_KERNEL_IS_PLAIN_JAVA = classes()
            .that().resideInAPackage("com.ridehailing.common.kernel..")
            .should().onlyDependOnClassesThat().resideInAnyPackage("java..", "com.ridehailing.common.kernel..");

    /**
     * Redis is covered by {@link #ONLY_THE_COMMON_CACHE_TALKS_TO_REDIS} and the domain by {@link #DOMAIN_IS_PLAIN_JAVA}.
     */
    static final ArchRule APPLICATION_AND_WEB_DO_NOT_USE_SQL_OR_KAFKA = noClasses()
            .that().resideInAnyPackage("..application..", "..adapters.web..")
            .should().dependOnClassesThat().resideInAnyPackage(
                    "org.springframework.jdbc..", "org.springframework.kafka..", "org.apache.kafka..")
            .as("The application and the web adapter use neither SQL nor Kafka");

    /**
     * Use cases refuse requests with {@code common.error} exceptions, which name what went wrong, not an HTTP status:
     * the web adapter's {@code ApiExceptionHandler} decides the status.
     */
    static final ArchRule APPLICATION_AND_ERRORS_KNOW_NOTHING_OF_HTTP = noClasses()
            .that().resideInAnyPackage("..application..", "com.ridehailing.common.error..")
            .should().dependOnClassesThat().resideInAnyPackage(
                    "org.springframework.http..", "org.springframework.web..", "jakarta.servlet..")
            .as("The application and the exceptions it throws know nothing of HTTP");

    /**
     * The outbox, de-duplication and event codec are interfaces; their JDBC and Jackson implementations hide.
     */
    static final ArchRule APPLICATION_USES_COMMON_EVENTS_THROUGH_INTERFACES = noClasses()
            .that().resideInAPackage("com.ridehailing.*.application..")
            .should().dependOnClassesThat(resideInAPackage("com.ridehailing.common.events..").and(not(INTERFACES)))
            .orShould().dependOnClassesThat().resideInAnyPackage("tools.jackson..", "com.fasterxml.jackson..")
            .as("The application uses common.events only through interfaces, and never Jackson");

    static final ArchRule CONTRACTS_ARE_PLAIN_JAVA = classes()
            .that().resideInAPackage("com.ridehailing.contracts..")
            .should().onlyDependOnClassesThat().resideInAnyPackage("java..", "com.ridehailing.contracts..");

    static final ArchRule ONLY_THE_OUTBOX_RELAY_PUBLISHES_TO_KAFKA = noClasses()
            .that().doNotHaveSimpleName("OutboxRelay")
            .and().doNotHaveSimpleName("MessagingConfig")
            .should().dependOnClassesThat().haveFullyQualifiedName("org.springframework.kafka.core.KafkaTemplate");

    static final ArchRule ONLY_THE_COMMON_CACHE_TALKS_TO_REDIS = noClasses()
            .that().resideOutsideOfPackage("com.ridehailing.common.cache..")
            .should().dependOnClassesThat().resideInAPackage("org.springframework.data.redis..");

    /**
     * Lombok only writes bean constructors and loggers. The web request and response models stay plain records;
     * lombok.config marks generated members with {@code @lombok.Generated}, which is how this rule sees it. The domain,
     * contracts and the shared kernel are covered by their plain-Java rules, which reject that annotation too.
     */
    static final ArchRule NO_LOMBOK_IN_WEB_MODELS = noMembers()
            .that().areDeclaredInClassesThat().resideInAPackage("..adapters.web.model..")
            .should().beAnnotatedWith("lombok.Generated")
            .as("No Lombok in the web request and response models");

    private ArchitectureRules() {
    }

    static List<ArchRule> all() {
        return List.of(DEPENDENCIES_POINT_INWARDS, ADAPTERS_DO_NOT_DEPEND_ON_EACH_OTHER, PORTS_ARE_INTERFACES,
                DOMAIN_IS_PLAIN_JAVA, SHARED_KERNEL_IS_PLAIN_JAVA, CONTRACTS_ARE_PLAIN_JAVA,
                APPLICATION_AND_ERRORS_KNOW_NOTHING_OF_HTTP, APPLICATION_USES_COMMON_EVENTS_THROUGH_INTERFACES,
                APPLICATION_AND_WEB_DO_NOT_USE_SQL_OR_KAFKA, ONLY_THE_OUTBOX_RELAY_PUBLISHES_TO_KAFKA,
                ONLY_THE_COMMON_CACHE_TALKS_TO_REDIS, NO_LOMBOK_IN_WEB_MODELS);
    }
}
