#!/usr/bin/env bash
# Start charging → provider → notification, then frontend, for local E2E / Postman testing.
set -u

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT" || exit 1

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
BLUE='\033[0;34m'
WHITE='\033[0;37m'
BOLD='\033[1m'
NC='\033[0m'

PIDS=()
CLEANING_UP=0

usage() {
  cat <<'EOF'
Usage: start-all.sh [--help]

  Start all Python/FastAPI backend services in dependency order, then the React frontend:
  charging-service (8003) → provider-service (8002) → notification-service (8001)
  → frontend dev server (4200).

  Requires a virtual environment at .venv/, Node/Yarn for the frontend, and curl for health checks.
  Press Ctrl+C to stop all services.

Options:
  --help    Show this help message.
EOF
}

if [[ "${1:-}" == "--help" || "${1:-}" == "-h" ]]; then
  usage
  exit 0
fi

prefix_pipe() {
  local tag=$1
  local color=$2
  while IFS= read -r line || [[ -n "${line}" ]]; do
    printf '%b[%s]%b %s\n' "${color}" "${tag}" "${NC}" "${line}"
  done
}

cleanup() {
  if [[ "${CLEANING_UP}" -eq 1 ]]; then
    return 0
  fi
  CLEANING_UP=1
  trap - INT TERM
  echo -e "\n${YELLOW}Stopping all services...${NC}"
  local pid
  for pid in "${PIDS[@]}"; do
    if kill -0 "${pid}" 2>/dev/null; then
      kill "${pid}" 2>/dev/null || true
    fi
  done
  sleep 1
  local port pids
  for port in 8001 8002 8003 4200; do
    pids=$(lsof -nP -iTCP:"${port}" -sTCP:LISTEN -t 2>/dev/null || true)
    if [[ -n "${pids}" ]]; then
      kill ${pids} 2>/dev/null || true
    fi
  done
  sleep 0.5
  for port in 8001 8002 8003 4200; do
    pids=$(lsof -nP -iTCP:"${port}" -sTCP:LISTEN -t 2>/dev/null || true)
    if [[ -n "${pids}" ]]; then
      kill -9 ${pids} 2>/dev/null || true
    fi
  done
  echo -e "${GREEN}Stopped.${NC}"
  exit 0
}

trap cleanup INT TERM

if [[ ! -f "${ROOT}/.venv/bin/activate" ]]; then
  echo -e "${RED}No virtual environment found at ${ROOT}/.venv${NC}" >&2
  echo "Create one with: python3 -m venv .venv && source .venv/bin/activate && pip install -r apps/notification-service/requirements.txt" >&2
  exit 1
fi

# shellcheck source=/dev/null
source "${ROOT}/.venv/bin/activate"

export PYTHONPATH="${ROOT}/libs/grpc-contracts:${ROOT}/libs/py-core${PYTHONPATH:+:${PYTHONPATH}}"
export PYTHONUNBUFFERED=1

# --- Postgres check & migrate ---
export DATABASE_URL="${DATABASE_URL:-postgresql://gondo:gondo@localhost:5432/gondo}"

echo -e "${YELLOW}Checking Postgres...${NC}"
if ! docker-compose ps postgres --format '{{.State}}' 2>/dev/null | grep -qi running; then
  echo -e "${YELLOW}Starting Postgres via docker-compose...${NC}"
  docker-compose up -d postgres
fi
wait_for_health_pg() {
  local attempt=0
  local max=60
  echo -e "${YELLOW}Waiting for Postgres health...${NC}"
  while ((attempt < max)); do
    if docker-compose exec -T postgres pg_isready -U gondo -d gondo >/dev/null 2>&1; then
      echo -e "${GREEN}Postgres is ready.${NC}"
      return 0
    fi
    attempt=$((attempt + 1))
    sleep 0.5
  done
  echo -e "${RED}Timeout waiting for Postgres.${NC}" >&2
  return 1
}
wait_for_health_pg || { cleanup; exit 1; }

echo -e "${YELLOW}Running database migrations...${NC}"
"${ROOT}/scripts/db-migrate.sh"
echo -e "${GREEN}Migrations complete.${NC}"
echo ""
# --- End Postgres check ---

echo -e "${BOLD}${GREEN}Workspace:${NC} ${ROOT}"
echo -e "${BOLD}${GREEN}PYTHONPATH:${NC} ${PYTHONPATH}"
echo ""

wait_for_health() {
  local port=$1
  local name=$2
  local url="http://127.0.0.1:${port}/health"
  local attempt=0
  local max=120
  echo -e "${YELLOW}Waiting for ${name} health (${url})...${NC}"
  while ((attempt < max)); do
    if curl -sf --connect-timeout 2 "${url}" >/dev/null 2>&1; then
      echo -e "${GREEN}${name} is healthy.${NC}"
      return 0
    fi
    attempt=$((attempt + 1))
    sleep 0.5
  done
  echo -e "${RED}Timeout waiting for ${name} on port ${port}.${NC}" >&2
  return 1
}

wait_for_frontend() {
  local attempt=0
  local max=120
  echo -e "${YELLOW}Waiting for frontend (http://localhost:4200/)...${NC}"
  while ((attempt < max)); do
    if curl -sf http://localhost:4200/ >/dev/null 2>&1; then
      echo -e "${GREEN}frontend is up.${NC}"
      return 0
    fi
    attempt=$((attempt + 1))
    sleep 0.5
  done
  echo -e "${RED}Timeout waiting for frontend on port 4200.${NC}" >&2
  return 1
}

start_frontend() {
  local tag=$1
  local color=$2
  (
    cd "${ROOT}" || exit 1
    exec yarn nx run frontend:serve
  ) > >(prefix_pipe "${tag}" "${color}") 2>&1 &
  PIDS+=($!)
}

start_uvicorn() {
  local tag=$1
  local color=$2
  local app_dir=$3
  local port=$4
  (
    cd "${ROOT}/${app_dir}" || exit 1
    export DATABASE_URL="${DATABASE_URL:-postgresql://gondo:gondo@localhost:5432/gondo}"
    exec python -m uvicorn main:app --reload --port "${port}"
  ) > >(prefix_pipe "${tag}" "${color}") 2>&1 &
  PIDS+=($!)
}

echo -e "${CYAN}[bootstrap]${NC} Starting charging-service (8003)..."
start_uvicorn "charging" "${CYAN}" "apps/charging-service" 8003
sleep 1
wait_for_health 8003 "charging-service" || {
  cleanup
  exit 1
}

echo -e "${MAGENTA}[bootstrap]${NC} Starting provider-service (8002)..."
start_uvicorn "provider" "${MAGENTA}" "apps/provider-service" 8002
sleep 1
wait_for_health 8002 "provider-service" || {
  cleanup
  exit 1
}

echo -e "${BLUE}[bootstrap]${NC} Starting notification-service (8001)..."
start_uvicorn "notification" "${BLUE}" "apps/notification-service" 8001
sleep 1
wait_for_health 8001 "notification-service" || {
  cleanup
  exit 1
}

echo -e "${WHITE}[bootstrap]${NC} Starting frontend (4200)..."
start_frontend "frontend" "${WHITE}"
sleep 1
wait_for_frontend || {
  cleanup
  exit 1
}

echo ""
echo -e "${BOLD}${GREEN}============================================${NC}"
echo -e "${BOLD}${GREEN}  All services are running!${NC}"
echo -e "${BOLD}${GREEN}============================================${NC}"
echo -e "  Frontend:                     http://localhost:4200"
echo -e "  Notification Service (main):  http://localhost:8001"
echo -e "  Provider Service:             http://localhost:8002"
echo -e "  Charging Service:             http://localhost:8003"
echo ""
echo -e "  Test endpoints:"
echo -e "    GET http://localhost:8001/health"
echo ""
echo -e "  Swagger UI (REST):"
echo -e "    http://localhost:8001/docs  (Notification)"
echo -e "    http://localhost:8002/docs  (Provider)"
echo -e "    http://localhost:8003/docs  (Charging)"
echo ""
echo -e "  gRPC Reflection (grpcurl / grpcui):"
echo -e "    Provider  :50051  ${CYAN}grpcui -plaintext localhost:50051${NC}"
echo -e "    Charging  :50052  ${CYAN}grpcui -plaintext localhost:50052${NC}"
echo ""
echo -e "  ${BOLD}Press Ctrl+C to stop all services.${NC}"
echo -e "${BOLD}${GREEN}============================================${NC}"
echo ""

wait
