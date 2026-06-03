#!/usr/bin/env bash
set -euo pipefail

PORT="${DEMO_PORT:-3000}"

if ! command -v cloudflared >/dev/null 2>&1; then
  echo "cloudflared não encontrado."
  echo "Arch:  sudo pacman -S cloudflared"
  echo "Debian: https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/downloads/"
  exit 1
fi

echo ""
echo "=== Cloudflare Quick Tunnel ==="
echo "Apontando para http://127.0.0.1:${PORT}"
echo ""
echo "Quando aparecer a URL https://....trycloudflare.com :"
echo "  1. Compartilhe esse link (demo pública)"
echo "  2. Configure MinIO (não sobrescreva deploy/.env com .env.example):"
echo "     make demo-set-minio URL=https://SEU_HOST.trycloudflare.com/storage"
echo "  3. make demo-restart-api"
echo ""
echo "Ctrl+C encerra o túnel."
echo ""

exec cloudflared tunnel --url "http://127.0.0.1:${PORT}"
