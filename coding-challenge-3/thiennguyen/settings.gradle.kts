plugins {
    id("org.gradle.toolchains.foojay-resolver-convention") version "1.0.0"
}

rootProject.name = "ride-hailing"

// One Gradle project per business module, so each can later be extracted into its own service. :app assembles them
// into today's modular monolith and owns the cross-module tests.
val modules = listOf("contracts", "common", "driver", "ride", "dispatch")
modules.forEach { module ->
    include(":$module")
    project(":$module").projectDir = file("modules/$module")
}
include(":app")
