plugins {
    `java-library`
    `java-test-fixtures`
}

dependencies {
    // Part of the public API of every module that depends on common (web, validation, JDBC, Jackson).
    api(libs.spring.boot.starter.webmvc)
    api(libs.spring.boot.starter.validation)
    api(libs.spring.boot.starter.jdbc)
    api(libs.spring.modulith.api)
    api(project(":contracts"))

    implementation(libs.spring.boot.starter.kafka)
    implementation(libs.spring.boot.starter.data.redis)
    implementation(libs.micrometer.core)
}
