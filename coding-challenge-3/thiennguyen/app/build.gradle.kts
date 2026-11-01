import java.time.Duration
import java.nio.file.Paths

// The deployable modular monolith: wires every module into one Spring Boot application and owns the tests that need
// all of them (module boundaries, architecture rules, acceptance tests against real infrastructure).
plugins {
    java
    id("org.springframework.boot")
}

dependencies {
    implementation(project(":contracts"))
    implementation(project(":common"))
    implementation(project(":driver"))
    implementation(project(":ride"))
    implementation(project(":dispatch"))

    implementation(libs.spring.boot.starter.actuator)
    implementation(libs.spring.boot.starter.flyway)
    implementation(libs.flyway.database.postgresql)
    implementation(libs.spring.boot.starter.kafka)
    implementation(libs.spring.boot.starter.data.redis)
    implementation(libs.spring.modulith.starter.core)
    runtimeOnly(libs.spring.modulith.actuator)
    runtimeOnly(libs.postgresql)
    developmentOnly(libs.spring.boot.docker.compose)

    testImplementation(libs.spring.modulith.starter.test)
    testImplementation(libs.archunit)
}

// Dynamically looks for the Colima Docker socket for any local user (target MacOS)
fun getColimaSocketPath(): String {
    val colimaSocket = Paths.get(System.getProperty("user.home"), ".colima", "default", "docker.sock").toFile()
    return if (colimaSocket.exists()) {
        "unix://${colimaSocket.absolutePath}"
    } else {
        System.getenv("DOCKER_HOST") ?: "unix:///var/run/docker.sock"
    }
}

testing {
    suites {
        val test by getting(JvmTestSuite::class)

        val integrationTest by registering(JvmTestSuite::class) {
            useJUnitJupiter()
            configurations[sources.implementationConfigurationName]
                .extendsFrom(configurations.testImplementation.get())
            configurations[sources.runtimeOnlyConfigurationName]
                .extendsFrom(configurations.testRuntimeOnly.get())
            dependencies {
                implementation(project())
                implementation(libs.spring.boot.testcontainers)
                implementation(libs.testcontainers.junit.jupiter)
                implementation(libs.testcontainers.postgresql)
                implementation(libs.testcontainers.kafka)
                implementation(libs.awaitility)
            }
            targets.all {
                testTask.configure {
                    shouldRunAfter(test)
                    timeout = Duration.ofMinutes(10)

                    // Inject the Colima socket path directly into the Integration Test runner process
                    environment("DOCKER_HOST", getColimaSocketPath())
                    // Colima stability performance variables for Testcontainers
                    environment("TESTCONTAINERS_CHECKS_DISABLE", "true")
                    environment("TESTCONTAINERS_RYUK_DISABLED", "true")
                    systemProperty("testcontainers.container.startup.timeout", "120")
                    systemProperty("container.creation.timeout", "120")
                }
            }
        }
    }
}

tasks.bootJar {
    archiveFileName = "app.jar"
}

// Run from the repository root so spring-boot-docker-compose finds compose.yaml and starts the infrastructure.
// Gradle's output is not a TTY, so Spring's ANSI detection would turn the console colours off; force them on here.
tasks.named<org.springframework.boot.gradle.tasks.run.BootRun>("bootRun") {
    workingDir = rootProject.projectDir
    systemProperty("spring.output.ansi.enabled", "always")
}
