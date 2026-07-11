#!/usr/bin/env bash
# One-time migration: seed watched items from the old Sale Watcher into Timeline.
#
# Usage:
#   TIMELINE_URL=https://timeline-production-a8a3.up.railway.app \
#   TIMELINE_PASSWORD=yourpassword \
#   ./scripts/migrate_sales.sh
#
# Floors are seeded at ~30% of each threshold to reject shipping-only / sub-value
# false positives. Adjust them afterward in the Sales UI as needed.
set -euo pipefail

BASE="${TIMELINE_URL:-http://localhost:8080}"
PASS="${TIMELINE_PASSWORD:-}"
JAR="$(mktemp)"

if [ -n "$PASS" ]; then
  echo "Logging in to $BASE ..."
  curl -s -c "$JAR" -X POST "$BASE/api/login" \
    -H 'Content-Type: application/json' \
    -d "{\"password\":\"$PASS\"}" >/dev/null
fi

create() {
  local name="$1" term="$2" threshold="$3" floor="$4"
  echo "Adding: $name (threshold \$$threshold, floor \$$floor)"
  curl -s -b "$JAR" -X POST "$BASE/api/sales/items" \
    -H 'Content-Type: application/json' \
    -d "{\"name\":\"$name\",\"search_term\":\"$term\",\"threshold\":$threshold,\"floor\":$floor,\"category\":\"NVMe\"}" \
    -w '  [%{http_code}]\n' -o /dev/null
}

create "nvme gen 3 2tb" "nvme gen 3 2tb" 200 60
create "nvme gen 3 1tb" "nvme gen 3 1tb" 100 30
create "nvme gen 3 4tb" "nvme gen 3 4tb" 350 100

rm -f "$JAR"
echo "Done."
