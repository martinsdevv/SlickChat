#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
ENV_FILE="$ROOT/deploy/.env"
URL="${1:-}"

if [[ -z "$URL" ]]; then
  echo "Uso: $0 'https://seu-host.trycloudflare.com/storage'" >&2
  exit 1
fi

URL="${URL%/}"

if [[ ! -f "$ENV_FILE" ]]; then
  if [[ -f "$ROOT/deploy/.env.example" ]]; then
    cp "$ROOT/deploy/.env.example" "$ENV_FILE"
  else
    echo "DEMO_PORT=3000" >"$ENV_FILE"
    echo "MINIO_PUBLIC_URL=" >>"$ENV_FILE"
  fi
fi

if grep -q '^MINIO_PUBLIC_URL=' "$ENV_FILE"; then
  sed -i "s|^MINIO_PUBLIC_URL=.*|MINIO_PUBLIC_URL=$URL|" "$ENV_FILE"
elif grep -q '^# MINIO_PUBLIC_URL=' "$ENV_FILE"; then
  sed -i "s|^# MINIO_PUBLIC_URL=.*|MINIO_PUBLIC_URL=$URL|" "$ENV_FILE"
else
  echo "MINIO_PUBLIC_URL=$URL" >>"$ENV_FILE"
fi

echo "deploy/.env atualizado:"
grep '^MINIO_PUBLIC_URL=' "$ENV_FILE" || true
echo "Reinicie a API: make demo-restart-api"
