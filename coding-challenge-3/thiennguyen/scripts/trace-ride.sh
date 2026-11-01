#!/usr/bin/env bash
# Prints every application log line for one ride, answering "what happened to ride X?" from the logs alone.
# Reads the app container's logs, so the stack must run in Docker (`make up`); `./gradlew :app:bootRun` logs to your
# terminal instead, and `make down` removes the container and its logs.
# Usage: scripts/trace-ride.sh <rideId>
set -euo pipefail
cd "$(dirname "$0")/.."
# Detect Docker Compose binary/command
detect_compose() {
    # 1. Check for modern Docker Compose V2 (plugin)
    if docker compose version >/dev/null 2>&1; then
        echo "docker compose"
    # 2. Check for legacy Docker Compose V1 (standalone)
    elif command -v docker-compose >/dev/null 2>&1; then
        echo "docker-compose"
    else
        echo ""
    fi
}

DOCKER_COMPOSE="$(detect_compose)"

ride_id="${1:?usage: trace-ride.sh <rideId>   (make demo prints one)}"
if [[ ! "$ride_id" =~ ^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$ ]]; then
  echo "'$ride_id' is not a ride id: ride ids are UUIDs, e.g. the id POST /rides returned" >&2
  exit 1
fi

logs="$($DOCKER_COMPOSE logs --no-log-prefix --no-color app 2>/dev/null || true)"
if [[ -z "$logs" ]]; then
  echo "No logs from the app container. Start the stack with 'make up' (trace reads its logs; it cannot see" >&2
  echo "'./gradlew :app:bootRun' output, and 'make down' removes them)." >&2
  exit 1
fi

# The app logs in colour (compose.yaml); keep it for a terminal, strip it when the output is piped or redirected.
if [[ ! -t 1 ]]; then
  logs="$(printf '%s\n' "$logs" | sed $'s/\e\\[[0-9;]*m//g')"
fi

# Every line names its ride in the logging pattern as "rideId=<id>" (CONSOLE_LOG_PATTERN in app/src/main/resources/logback-spring.xml).
if ! printf '%s\n' "$logs" | grep -F "rideId=$ride_id"; then
  echo "No log lines for ride $ride_id since the app container started. Check the id with: curl -s localhost:8080/rides/$ride_id" >&2
  exit 1
fi
