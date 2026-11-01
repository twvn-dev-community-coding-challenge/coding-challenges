plugins {
    `java-library`
}

dependencies {
    api(project(":common"))
    implementation(libs.spring.boot.starter.kafka)
    implementation(libs.micrometer.core)

    testImplementation(libs.spring.boot.starter.webmvc.test)
}
