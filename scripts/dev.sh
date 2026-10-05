#!/usr/bin/env bash
# Start the Go API and the React dev server. The UI is http://localhost:3000
# and API calls stay on that origin; Create React App proxies them to the API,
# which is how the built UI is served in production (one origin).
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"

api_port=9090
pids=()

cleanup() {
  local pid
  for pid in "${pids[@]:-}"; do
    if [[ -n "${pid}" ]] && kill -0 "${pid}" 2>/dev/null; then
      kill -- "-${pid}" 2>/dev/null || kill "${pid}" 2>/dev/null || true
    fi
  done
}
trap cleanup EXIT INT TERM

set -m

echo "API  http://127.0.0.1:${api_port}"
go run ./cmd/server &
pids+=("$!")

ready=0
for _ in $(seq 1 60); do
  if (echo >/dev/tcp/127.0.0.1/"${api_port}") >/dev/null 2>&1; then
    ready=1
    break
  fi
  if ! kill -0 "${pids[0]}" 2>/dev/null; then
    echo "API exited before it started listening"
    exit 1
  fi
  sleep 0.5
done
if [[ "${ready}" -ne 1 ]]; then
  echo "API did not open :${api_port}; starting the UI anyway"
fi

echo "UI   http://localhost:3000  (API proxied to :${api_port})"
(
  cd "${root}/web"
  export BROWSER=none
  export REACT_APP_API_BASE_URL=
  exec npm start
) &
pids+=("$!")

wait -n
status=$?
cleanup
exit "${status}"
