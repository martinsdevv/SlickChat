#!/usr/bin/env bash
# Carrega deploy/.env sem alterar o arquivo (somente leitura).
# Uso: source deploy/scripts/load-env.sh

_DEPLOY_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
_ENV_FILE="$_DEPLOY_DIR/.env"

if [[ ! -f "$_ENV_FILE" ]]; then
  return 0 2>/dev/null || exit 0
fi

set -a
# shellcheck disable=SC1090
source "$_ENV_FILE"
set +a
