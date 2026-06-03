#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT/deploy"

if [[ ! -d ../frontend/web/dist ]]; then
  echo "Build do frontend ausente. Rode: make demo-build"
  exit 1
fi

docker compose -f compose.demo.yml up -d
echo "Proxy em http://127.0.0.1:${DEMO_PORT:-3000}"
