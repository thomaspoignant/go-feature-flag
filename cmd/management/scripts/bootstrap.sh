#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICE_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

: "${GOFF_MGMT_DB_URL:=postgres://goff:goff@localhost:5432/goff_mgmt?sslmode=disable}"

echo "==> Waiting for PostgreSQL"
for i in {1..30}; do
  if PGPASSWORD=goff psql -h localhost -U goff -d goff_mgmt -c 'SELECT 1' >/dev/null 2>&1; then
    echo "    postgres is up"
    break
  fi
  sleep 1
  if [ "$i" = "30" ]; then
    echo "    postgres not reachable after 30s" >&2
    exit 1
  fi
done

echo "==> Applying goose migrations"
cd "${SERVICE_DIR}"
goose -dir ./db/migrations postgres "${GOFF_MGMT_DB_URL}" up

if [ -f "${SCRIPT_DIR}/seed.sql" ] && [ "${GOFF_MGMT_SEED:-1}" = "1" ]; then
  echo "==> Seeding (scripts/seed.sql)"
  PGPASSWORD=goff psql -h localhost -U goff -d goff_mgmt -f "${SCRIPT_DIR}/seed.sql"
fi

echo "==> Done."
