#!/usr/bin/env bash
# Waits until the API reports healthy, so `make up` returns only when `make demo` can run.
# Usage: scripts/wait-for-health.sh [timeout-seconds]   (BASE_URL defaults to http://localhost:8080)
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
timeout="${1:-180}"
deadline=$((SECONDS + timeout))

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

printf 'Waiting for %s/actuator/health ' "$BASE_URL"
until curl -fsS "$BASE_URL/actuator/health" 2>/dev/null | grep -q '"status":"UP"'; do
  if (( SECONDS >= deadline )); then
    echo
    echo "The app did not become healthy within ${timeout}s. Recent app logs:" >&2
    $DOCKER_COMPOSE logs --no-log-prefix --tail=30 app >&2 || true
    echo "Follow the logs with: make logs" >&2
    exit 1
  fi
  printf '.'
  sleep 2
done
echo " UP"
