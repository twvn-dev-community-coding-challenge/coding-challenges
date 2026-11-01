#!/usr/bin/env bash
# Walks one ride through create -> offer -> reject -> rematch -> accept -> complete against a running stack
# (`make up`), using the seeded passenger P001 and drivers near the brief's example pickup.
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"

if ! curl -fsS "$BASE_URL/actuator/health" 2>/dev/null | grep -q '"status":"UP"'; then
  echo "The API is not healthy at $BASE_URL. Start the stack first: make up" >&2
  exit 1
fi

field() { sed -n "s/.*\"$1\":\"\{0,1\}\([^\",}]*\)\"\{0,1\}.*/\1/p" | head -1; }

wait_for_offer() {
  local ride_id="$1" previous="${2:-}"
  for _ in $(seq 1 40); do
    local body status driver
    body="$(curl -fsS "$BASE_URL/rides/$ride_id")"
    status="$(printf '%s' "$body" | field status)"
    driver="$(printf '%s' "$body" | field assignedDriverId)"
    if [[ "$status" == "DRIVER_ASSIGNED" && "$driver" != "$previous" && "$driver" != "null" ]]; then
      printf '%s' "$driver"
      return 0
    fi
    sleep 0.5
  done
  echo "Ride $ride_id was not offered to a driver in time" >&2
  return 1
}

echo "1. Passenger P001 requests a ride"
created="$(curl -fsS -X POST "$BASE_URL/rides" \
  -H 'Content-Type: application/json' -H "Idempotency-Key: demo-$(date +%s)-$RANDOM" \
  -d '{"passengerId":"P001","pickup":{"latitude":10.7769,"longitude":106.7009},"destination":{"latitude":10.8231,"longitude":106.6297}}')"
ride_id="$(printf '%s' "$created" | field id)"
echo "   ride $ride_id created: $(printf '%s' "$created" | field status)"

first="$(wait_for_offer "$ride_id")"
echo "2. Offered to the nearest available driver: $first"

curl -fsS -X POST "$BASE_URL/rides/$ride_id/offers/$first/reject" > /dev/null
echo "3. $first rejects; the ride returns to MATCHING and $first is excluded from it"

second="$(wait_for_offer "$ride_id" "$first")"
echo "4. Rematched to the next nearest driver: $second"

accepted="$(curl -fsS -X POST "$BASE_URL/rides/$ride_id/offers/$second/accept")"
echo "5. $second accepts; the offer is $(printf '%s' "$accepted" | field status)"
for _ in $(seq 1 40); do
  ride_status="$(curl -fsS "$BASE_URL/rides/$ride_id" | field status)"
  [[ "$ride_status" == "DRIVER_ACCEPTED" ]] && break
  sleep 0.25
done
echo "   ride is now $ride_status (updated from dispatch's RideAccepted event)"

completed="$(curl -fsS -X POST "$BASE_URL/rides/$ride_id/complete")"
echo "6. Trip completed: $(printf '%s' "$completed" | field status)"

echo
echo "History:   curl -s $BASE_URL/rides/$ride_id/history"
echo "Offers:    curl -s $BASE_URL/rides/$ride_id/offers"
echo "Logs:      make trace RIDE=$ride_id"
