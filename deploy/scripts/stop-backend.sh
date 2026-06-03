#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"

bash "$ROOT/deploy/scripts/kill-port.sh" 8081
bash "$ROOT/deploy/scripts/kill-port.sh" 8080

for name in api gateway fanout persistence ttl; do
  pidfile="deploy/logs/${name}.pid"
  if [[ -f "$pidfile" ]]; then
    pid=$(cat "$pidfile")
    if kill -0 "$pid" 2>/dev/null; then
      kill "$pid" 2>/dev/null || true
      echo "[$name] encerrado (pid $pid)"
    fi
    rm -f "$pidfile"
  fi
done

pkill -f "./services/api/cmd" 2>/dev/null || true
pkill -f "./services/gateway/cmd" 2>/dev/null || true
pkill -f "./services/workers/fanout/cmd" 2>/dev/null || true
pkill -f "./services/workers/persistence/cmd" 2>/dev/null || true
pkill -f "./services/workers/ttl/cmd" 2>/dev/null || true
