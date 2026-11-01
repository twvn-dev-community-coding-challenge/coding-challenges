// Dispatch depends only on common and the published contracts, never on ride or driver, so it can be extracted into
// its own service. Its package-info declares the same allowed dependencies; DispatchModuleIT boots it on its own.
plugins {
    `java-library`
}

dependencies {
    api(project(":common"))
    implementation(project(":contracts"))
    implementation(libs.spring.boot.starter.kafka)
    implementation(libs.micrometer.core)

    testImplementation(testFixtures(project(":common")))
}
