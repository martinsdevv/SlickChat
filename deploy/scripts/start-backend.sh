#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"
mkdir -p deploy/logs

# shellcheck disable=SC1091
source "$ROOT/deploy/scripts/load-env.sh"

start_one() {
  local name=$1
  shift
  local pattern=$1
  if pgrep -f "$pattern" >/dev/null 2>&1; then
    echo "[$name] já em execução"
    return
  fi
  nohup go run "$@" >"deploy/logs/${name}.log" 2>&1 &
  echo $! >"deploy/logs/${name}.pid"
  echo "[$name] pid $(cat "deploy/logs/${name}.pid") — deploy/logs/${name}.log"
}

echo "Iniciando serviços Go..."
start_one api "./services/api/cmd"
start_one gateway "./services/gateway/cmd"
start_one fanout "./services/workers/fanout/cmd"
start_one persistence "./services/workers/persistence/cmd"
start_one ttl "./services/workers/ttl/cmd"
echo "Aguardando API (8081)..."
for _ in $(seq 1 60); do
  code=$(curl -s -o /dev/null -w "%{http_code}" "http://127.0.0.1:8081/login" \
    -X POST -H "Content-Type: application/json" -d "{}" 2>/dev/null || echo "000")
  if [[ "$code" != "000" ]]; then
    echo "Backend pronto."
    exit 0
  fi
  sleep 0.5
done
echo "API ainda não respondeu — veja deploy/logs/api.log"
exit 1
