# syntax=docker/dockerfile:1.7

# Build stage: uses the Gradle image directly, so building the image needs only Docker.
FROM gradle:9.2.1-jdk25 AS build
WORKDIR /workspace
COPY settings.gradle.kts build.gradle.kts lombok.config ./
COPY gradle/libs.versions.toml gradle/
COPY modules modules
COPY app app
RUN --mount=type=cache,target=/home/gradle/.gradle \
    gradle :app:bootJar --no-daemon --quiet \
 && java -Djarmode=tools -jar app/build/libs/app.jar extract --layers --launcher --destination build/extracted

# Runtime stage: JRE only, non-root, layered for cache-friendly rebuilds.
FROM eclipse-temurin:25-jre
RUN useradd --system --uid 10001 app
WORKDIR /app
COPY --from=build /workspace/build/extracted/dependencies/ ./
COPY --from=build /workspace/build/extracted/spring-boot-loader/ ./
COPY --from=build /workspace/build/extracted/snapshot-dependencies/ ./
COPY --from=build /workspace/build/extracted/application/ ./
USER app
EXPOSE 8080
ENTRYPOINT ["java", "-XX:MaxRAMPercentage=75", "org.springframework.boot.loader.launch.JarLauncher"]
