# -------------------------------------------------------------
# Shell Configuration & Defaults
# -------------------------------------------------------------
SHELL := /bin/sh
.SHELLFLAGS := -ec
.DEFAULT_GOAL := help

# -------------------------------------------------------------
# JDK 25 Discovery & Validation
# -------------------------------------------------------------
# If JAVA_HOME is passed explicitly via CLI (e.g. make JAVA_HOME=/path), use it directly.
# Otherwise, loop through candidates and select the first one whose binary actually reports Java 25.
ifeq ($(origin JAVA_HOME),command line)
    RESOLVED_JAVA_HOME := $(JAVA_HOME)
else
    RESOLVED_JAVA_HOME := $(shell \
        for candidate in \
            "$$JAVA_HOME" \
            "$$(/usr/libexec/java_home -v 25 2>/dev/null)" \
            "/opt/homebrew/opt/openjdk@25/libexec/openjdk.jdk/Contents/Home" \
            /opt/homebrew/Cellar/openjdk@25/*/libexec/openjdk.jdk/Contents/Home \
            /opt/homebrew/Cellar/openjdk/25*/libexec/openjdk.jdk/Contents/Home \
            /Library/Java/JavaVirtualMachines/*25*/Contents/Home \
            "$$HOME/.sdkman/candidates/java/25*" \
            /usr/lib/jvm/java-25* \
            /usr/lib/jvm/jdk-25*; do \
            if [ -n "$$candidate" ] && [ -x "$$candidate/bin/java" ]; then \
                if "$$candidate/bin/java" -version 2>&1 | grep -qE 'version "(25|25\.)'; then \
                    echo "$$candidate"; \
                    exit 0; \
                fi; \
            fi; \
        done)
endif

JAVA_HOME := $(RESOLVED_JAVA_HOME)
export JAVA_HOME

# -------------------------------------------------------------
# Docker Compose Detection
# -------------------------------------------------------------
DOCKER_COMPOSE ?= $(shell \
	if docker compose version >/dev/null 2>&1; then \
		echo "docker compose"; \
	elif command -v docker-compose >/dev/null 2>&1; then \
		echo "docker-compose"; \
	else \
		echo ""; \
	fi)

# -------------------------------------------------------------
# Phony Targets
# -------------------------------------------------------------
.PHONY: help setup jdk verify-docker wrapper images pull build test it acceptance ac verify clean up down logs demo trace

help: ## List available targets
	@echo "Usage: make <target>   (first time setup: make setup)"
	@grep -E '^[a-zA-Z_-]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  %-14s %s\n", $$1, $$2}'

verify-docker: ## Ensure Docker Compose is available
	@if [ -z "$(DOCKER_COMPOSE)" ]; then \
		echo "Error: Neither 'docker compose' nor 'docker-compose' was found on your PATH." >&2; \
		echo "Please install Docker Desktop or the Compose plugin." >&2; \
		exit 1; \
	fi

# -------------------------------------------------------------
# Verification Target
# -------------------------------------------------------------

jdk: ## Check that a JDK 25 is available for Gradle
	@if [ -z "$(JAVA_HOME)" ]; then \
		echo "Error: JDK 25 was not found on your system." >&2; \
		echo "Please install it with: brew install openjdk@25" >&2; \
		echo "Or specify its path manually: make JAVA_HOME=/path/to/jdk-25 <target>" >&2; \
		exit 1; \
	fi
	@echo "Using JDK 25 at $(JAVA_HOME)"

# The wrapper is committed to the repo; only triggers wrapper target if gradlew is missing
gradlew:
	$(MAKE) wrapper

wrapper: ## (Re)generate the Gradle wrapper via Docker without a local Gradle install
	docker run --rm -u "$$(id -u):$$(id -g)" -e GRADLE_USER_HOME=/tmp/gradle-home \
		-v "$(CURDIR)":/workspace -w /workspace gradle:9.2.1-jdk25 \
		gradle wrapper --gradle-version 9.2.1 --no-daemon

setup: jdk gradlew verify-docker images ## One-time setup: check JDK 25, download dependencies, pull Docker images
	./gradlew --quiet assemble testClasses :app:integrationTestClasses
	@echo "Setup complete. Next: make build (no Docker) or make up && make demo"

images: verify-docker ## Pull infrastructure images only if missing (reads tags from compose.yaml)
	@missing="$$(for img in $$($(DOCKER_COMPOSE) config --images); do docker image inspect "$$img" >/dev/null 2>&1 || echo "$$img"; done)"; \
	if [ -n "$$missing" ]; then \
		echo "Pulling missing images: $$missing"; \
		for img in $$missing; do docker pull "$$img"; done; \
	else \
		echo "All required infrastructure images are present locally"; \
	fi

pull: verify-docker ## Force-pull infrastructure images to refresh them
	$(DOCKER_COMPOSE) pull

build: jdk gradlew ## Compile all modules, run unit tests, and build the boot JAR
	./gradlew build

test: jdk gradlew ## Run unit and architecture tests only
	./gradlew test

it: jdk gradlew verify-docker images ## Run all integration tests (Testcontainers)
	./gradlew :app:integrationTest

acceptance: jdk gradlew verify-docker images ## Run acceptance tests only (com.ridehailing.acceptance.*)
	./gradlew :app:integrationTest --tests 'com.ridehailing.acceptance.*'

ac: jdk gradlew verify-docker images ## Run a specific acceptance criterion: make ac AC=09
	@test -n "$(AC)" || { echo "Usage: make ac AC=09"; exit 1; }
	./gradlew :app:integrationTest --tests 'com.ridehailing.acceptance.AC$(AC)*'

verify: build it ## Full CI pipeline: build + unit tests + integration tests

clean: ## Remove build outputs while preserving .gradle caches and wrappers
	-@./gradlew --stop >/dev/null 2>&1 || true
	find . -type d \( -name build -o -name .kotlin \) ! -path './.git/*' ! -path './.gradle/*' -prune -exec rm -rf {} +

up: verify-docker ## Build images, start app and infrastructure, and wait for health check
	$(DOCKER_COMPOSE) --profile app up --build -d
	@./scripts/wait-for-health.sh

down: verify-docker ## Stop and tear down containers, networks, and persistent volumes
	$(DOCKER_COMPOSE) --profile app down -v

logs: verify-docker ## Tail the application log output
	$(DOCKER_COMPOSE) logs -f app

demo: ## Execute the end-to-end ride simulation script
	./scripts/demo.sh

trace: ## Display the lifecycle trace of a single ride: make trace RIDE=<rideId>
	@test -n "$(RIDE)" || { echo "Usage: make trace RIDE=<rideId>   (tip: run 'make demo' to obtain an ID)"; exit 1; }
	./scripts/trace-ride.sh $(RIDE)