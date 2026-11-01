import io.spring.gradle.dependencymanagement.dsl.DependencyManagementExtension
import org.gradle.api.tasks.testing.logging.TestExceptionFormat
import org.springframework.boot.gradle.plugin.SpringBootPlugin

plugins {
    alias(libs.plugins.spring.boot) apply false
    alias(libs.plugins.spring.dependency.management) apply false
}

subprojects {
    group = "com.ridehailing"
    version = "0.1.0-SNAPSHOT"

    repositories {
        mavenCentral()
    }

    pluginManager.withPlugin("java") {
        apply(plugin = "io.spring.dependency-management")
        apply(plugin = "jacoco")

        extensions.configure<JavaPluginExtension> {
            toolchain.languageVersion = JavaLanguageVersion.of(libs.versions.java.get())
        }

        extensions.configure<DependencyManagementExtension> {
            imports {
                mavenBom(SpringBootPlugin.BOM_COORDINATES)
                mavenBom("org.springframework.modulith:spring-modulith-bom:${libs.versions.spring.modulith.get()}")
            }
        }

        // Spring resolves @PathVariable / @RequestParam names from bytecode. The Spring Boot plugin sets this flag,
        // but only in :app, so every module needs it explicitly.
        tasks.withType<JavaCompile>().configureEach {
            options.compilerArgs.add("-parameters")
        }

        extensions.configure<JacocoPluginExtension> {
            toolVersion = libs.versions.jacoco.get()
        }

        dependencies {
            // Constructors and loggers of Spring beans only; lombok.config and ModularityTest keep it out of entities,
            // models and contracts.
            "compileOnly"(libs.lombok)
            "annotationProcessor"(libs.lombok)
            "testImplementation"(libs.spring.boot.starter.test)
            "testRuntimeOnly"(libs.junit.platform.launcher)
        }

        tasks.withType<Test>().configureEach {
            useJUnitPlatform()
            testLogging {
                events("started", "passed", "skipped", "failed")
                exceptionFormat = TestExceptionFormat.FULL
            }
        }
    }
}
