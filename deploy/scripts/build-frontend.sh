#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT/frontend/web"

echo "Instalando dependências do frontend..."
npm ci

echo "Build de produção (API/WS relativos ao mesmo host)..."
VITE_API_BASE_URL=/api VITE_WS_BASE_URL=/ws npm run build

echo "Frontend em frontend/web/dist"
