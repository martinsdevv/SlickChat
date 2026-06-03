#!/usr/bin/env bash
# Libera uma porta TCP no host (Linux).
set -euo pipefail

PORT="${1:-}"
if [[ -z "$PORT" ]]; then
  echo "Uso: $0 <porta>" >&2
  exit 1
fi

if command -v fuser >/dev/null 2>&1; then
  fuser -k "${PORT}/tcp" 2>/dev/null || true
elif command -v ss >/dev/null 2>&1; then
  PIDS=$(ss -lptn "sport = :${PORT}" 2>/dev/null | grep -oP 'pid=\K[0-9]+' || true)
  if [[ -n "$PIDS" ]]; then
    # shellcheck disable=SC2086
    kill $PIDS 2>/dev/null || true
  fi
fi

sleep 0.3
