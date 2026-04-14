#!/usr/bin/env bash
# Run Alembic migrations for all service schemas.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT" || exit 1

GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m'

export DATABASE_URL="${DATABASE_URL:-postgresql://gondo:gondo@localhost:5432/gondo}"

if [[ ! -f "${ROOT}/.venv/bin/activate" ]]; then
  echo -e "${RED}No virtual environment found at ${ROOT}/.venv${NC}" >&2
  exit 1
fi

# shellcheck source=/dev/null
source "${ROOT}/.venv/bin/activate"

for svc in provider-service charging-service notification-service; do
  svc_dir="${ROOT}/apps/${svc}"
  if [[ -d "${svc_dir}/alembic" ]]; then
    echo -e "${YELLOW}Migrating ${svc}...${NC}"
    (cd "${svc_dir}" && python -m alembic upgrade head)
    echo -e "${GREEN}${svc} migrations applied.${NC}"
  else
    echo -e "${YELLOW}No alembic/ directory for ${svc}, skipping.${NC}"
  fi
done

echo -e "${GREEN}All migrations complete.${NC}"
