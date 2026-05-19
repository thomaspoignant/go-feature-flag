#!/usr/bin/env bash
# Launch the GO Feature Flag Management API for local development.
#
# Steps:
#   1. Load environment variables from cmd/management/.env (or .env.example)
#   2. Start Docker infrastructure (Postgres + Keycloak)
#   3. Wait for Postgres to be reachable
#   4. Apply goose migrations
#   5. Run the API via `go run`
#
# Flags:
#   --no-infra        skip "docker compose up -d"
#   --no-migrate      skip goose migrations
#   --rebuild-infra   docker compose down -v before bringing infra up
#   --build           build a binary to ./out/management instead of `go run`
#
# Requires: docker, goose, go.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICE_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${SERVICE_DIR}"

# --- options ---
RUN_INFRA=1
RUN_MIGRATE=1
REBUILD_INFRA=0
BUILD_ONLY=0
for arg in "$@"; do
  case "$arg" in
    --no-infra)      RUN_INFRA=0 ;;
    --no-migrate)    RUN_MIGRATE=0 ;;
    --rebuild-infra) REBUILD_INFRA=1 ;;
    --build)         BUILD_ONLY=1 ;;
    -h|--help)
      sed -n '1,20p' "$0" | sed 's/^# //;s/^#//'
      exit 0
      ;;
    *) echo "Unknown flag: $arg" >&2; exit 2 ;;
  esac
done

# --- env ---
ENV_FILE=""
if [ -f "${SERVICE_DIR}/.env" ]; then
  ENV_FILE="${SERVICE_DIR}/.env"
elif [ -f "${SERVICE_DIR}/.env.example" ]; then
  ENV_FILE="${SERVICE_DIR}/.env.example"
  echo "==> Using .env.example (copy to .env to customize)"
fi
if [ -n "${ENV_FILE}" ]; then
  echo "==> Loading env from ${ENV_FILE}"
  set -a
  # shellcheck disable=SC1090
  source "${ENV_FILE}"
  set +a
fi

: "${GOFF_MGMT_DB_URL:=postgres://goff:goff@localhost:5432/goff_mgmt?sslmode=disable}"

# --- tool checks ---
need() { command -v "$1" >/dev/null 2>&1 || { echo "Missing required tool: $1" >&2; exit 1; }; }
need go
[ "$RUN_INFRA" = "1" ] && need docker
[ "$RUN_MIGRATE" = "1" ] && need goose

# --- infra ---
if [ "$RUN_INFRA" = "1" ]; then
  if [ "$REBUILD_INFRA" = "1" ]; then
    echo "==> Rebuilding docker infra (down -v)"
    docker compose down -v
  fi
  echo "==> Starting docker infra (postgres + keycloak)"
  docker compose up -d
fi

# --- wait for postgres ---
if [ "$RUN_MIGRATE" = "1" ] || [ "$RUN_INFRA" = "1" ]; then
  echo "==> Waiting for Postgres"
  for i in {1..60}; do
    if docker compose exec -T postgres pg_isready -U goff -d goff_mgmt >/dev/null 2>&1; then
      echo "    postgres is up"
      break
    fi
    sleep 1
    if [ "$i" = "60" ]; then
      echo "    postgres not reachable after 60s" >&2
      exit 1
    fi
  done
fi

# --- migrations ---
if [ "$RUN_MIGRATE" = "1" ]; then
  echo "==> Applying goose migrations"
  goose -dir ./db/migrations postgres "${GOFF_MGMT_DB_URL}" up
fi

# --- run ---
echo "==> Starting API on :${GOFF_MGMT_SERVER_PORT:-8080}"
echo "    Swagger:  http://goff.local:${GOFF_MGMT_SERVER_PORT:-8080}/swagger/index.html"
echo "    Login:    http://goff.local:${GOFF_MGMT_SERVER_PORT:-8080}/auth/login"
echo

if [ "$BUILD_ONLY" = "1" ]; then
  mkdir -p out
  go build -o out/management .
  exec ./out/management
fi

exec go run .
