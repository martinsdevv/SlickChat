#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"
mkdir -p deploy/logs

# shellcheck disable=SC1091
source "$ROOT/deploy/scripts/load-env.sh"

bash "$ROOT/deploy/scripts/kill-port.sh" 8081
pkill -f "services/api/cmd" 2>/dev/null || true
sleep 0.5

if ss -ltn 2>/dev/null | grep -q ':8081 '; then
  echo "[api] ERRO: porta 8081 ainda em uso. Rode: make demo-stop-api" >&2
  exit 1
fi

nohup go run ./services/api/cmd >deploy/logs/api.log 2>&1 &
echo $! >deploy/logs/api.pid

echo "[api] aguardando porta 8081..."
for _ in $(seq 1 40); do
  code=$(curl -s -o /dev/null -w "%{http_code}" -X PUT \
    "http://127.0.0.1:8081/media/upload-put?object_key=probe" \
    -H "Content-Type: image/png" 2>/dev/null || echo "000")
  if [[ "$code" == "401" || "$code" == "400" || "$code" == "403" ]]; then
    echo "[api] pid $(cat deploy/logs/api.pid) — rota upload-put OK (http $code)"
    if [[ -n "${MINIO_PUBLIC_URL:-}" ]]; then
      echo "[api] MINIO_PUBLIC_URL=$MINIO_PUBLIC_URL"
    fi
    exit 0
  fi
  if [[ "$code" == "404" ]]; then
    echo "[api] ainda 404 em upload-put — veja deploy/logs/api.log" >&2
    tail -5 deploy/logs/api.log >&2 || true
    exit 1
  fi
  sleep 0.25
done

echo "[api] timeout — veja deploy/logs/api.log" >&2
exit 1
