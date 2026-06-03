#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"

run_psql() {
  docker exec -i slickchat-postgres psql -U postgres -d slickchat -v ON_ERROR_STOP=1 "$@"
}

if ! docker ps --format '{{.Names}}' | grep -qx slickchat-postgres; then
  echo "Postgres não está rodando. Execute: make infra-up"
  exit 1
fi

echo "Preparando registro de migrations..."
run_psql <<'SQL' >/dev/null
CREATE TABLE IF NOT EXISTS schema_migrations (
    filename TEXT PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
SQL

applied_count=$(
  docker exec slickchat-postgres psql -U postgres -d slickchat -tAc \
    "SELECT COUNT(*)::int FROM schema_migrations"
)
has_messages=$(
  docker exec slickchat-postgres psql -U postgres -d slickchat -tAc \
    "SELECT EXISTS (
       SELECT 1 FROM information_schema.tables
       WHERE table_schema = 'public' AND table_name = 'messages'
     )"
)

legacy=false
if [[ "$applied_count" == "0" && "$has_messages" == "t" ]]; then
  legacy=true
  echo "Banco já existente detectado — pulando migrations antigas já materializadas."
fi

migration_applied() {
  local fname=$1
  docker exec slickchat-postgres psql -U postgres -d slickchat -tAc \
    "SELECT 1 FROM schema_migrations WHERE filename = '$fname' LIMIT 1" | grep -q 1
}

mark_applied() {
  local fname=$1
  docker exec slickchat-postgres psql -U postgres -d slickchat -tAc \
    "INSERT INTO schema_migrations (filename) VALUES ('$fname') ON CONFLICT DO NOTHING" >/dev/null
}

echo "Aplicando migrations em slickchat-postgres..."
for file in "$ROOT"/infrastructure/postgres/migrations/*.sql; do
  fname=$(basename "$file")
  version="${fname%%_*}"

  if migration_applied "$fname"; then
    echo "  ⊘ $fname (já aplicada)"
    continue
  fi

  if [[ "$legacy" == true && $((10#$version)) -le 7 ]]; then
    echo "  ⊘ $fname (legado)"
    mark_applied "$fname"
    continue
  fi

  echo "  → $fname"
  run_psql <"$file" >/dev/null
  mark_applied "$fname"
done

echo "Migrations OK."
