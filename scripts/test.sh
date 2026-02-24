#!/usr/bin/env bash
set -euo pipefail

profile_file="$(mktemp)"
trap 'rm -f "$profile_file"' EXIT

go test -coverprofile="$profile_file" ./...

total_cov="$(go tool cover -func="$profile_file" | awk '/^total:/ {gsub("%", "", $3); print $3}')"
if [ -z "$total_cov" ]; then
  echo "Failed to parse total coverage"
  exit 1
fi

echo "Total coverage: ${total_cov}%"
awk -v cov="$total_cov" 'BEGIN { if (cov <= 80.0) { exit 1 } }' || {
  echo "Coverage check failed: expected total coverage > 80.0%"
  exit 1
}
