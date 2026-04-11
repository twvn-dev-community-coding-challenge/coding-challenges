#!/usr/bin/env bash
# Stop uvicorn and frontend dev server listeners on local ports (macOS-friendly).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT" || exit 1

YELLOW='\033[0;33m'
GREEN='\033[0;32m'
NC='\033[0m'

any=0
for port in 8001 8002 8003 4200; do
  pids=$(lsof -nP -iTCP:"${port}" -sTCP:LISTEN -t 2>/dev/null || true)
  if [[ -n "${pids}" ]]; then
    any=1
    echo -e "${YELLOW}Stopping listener(s) on port ${port}: ${pids}${NC}"
    # shellcheck disable=SC2086
    kill ${pids} 2>/dev/null || true
  fi
done

sleep 1

for port in 8001 8002 8003 4200; do
  pids=$(lsof -nP -iTCP:"${port}" -sTCP:LISTEN -t 2>/dev/null || true)
  if [[ -n "${pids}" ]]; then
    echo -e "${YELLOW}Force killing remaining on port ${port}: ${pids}${NC}"
    # shellcheck disable=SC2086
    kill -9 ${pids} 2>/dev/null || true
  fi
done

if [[ "${any}" -eq 0 ]]; then
  echo "No listeners found on ports 8001, 8002, 8003, 4200."
else
  echo -e "${GREEN}Done.${NC}"
fi
