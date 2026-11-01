plugins {
    `java-library`
}

dependencies {
    api(project(":common"))
    implementation(libs.spring.boot.starter.kafka)
}
