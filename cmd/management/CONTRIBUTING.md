# Contributing — GO Feature Flag Management API

This directory hosts the management API microservice for GO Feature Flag.
It is a separate Go module: source of truth for **Teams → Flagsets → Flags**
with OIDC login, audit log, and per-flag version history.

> Constraint: this module depends only on `github.com/thomaspoignant/go-feature-flag/modules/core` from the monorepo. Reuse from any other in-repo package is not allowed.

---

## 1. Prerequisites

| Tool | Version | Install |
|------|---------|---------|
| Go | 1.26+ | https://go.dev/dl/ |
| Docker + Compose | latest | https://docs.docker.com/get-docker/ |
| `make` | any | system package manager |
| `goose` | v3+ | `go install github.com/pressly/goose/v3/cmd/goose@latest` |
| `swag` | v1.16+ | `go install github.com/swaggo/swag/cmd/swag@latest` |
| `psql` | 14+ | for the bootstrap script (optional but recommended) |

---

## 2. One-time repo setup

From the **repo root**:

```bash
make workspace-init   # adds cmd/management to go.work
make vendor           # tidies + vendors all modules
```

From this directory (`cmd/management`):

```bash
go mod tidy           # first time: downloads our deps
cp .env.example .env  # then edit if needed
```

---

## 3. Start local infrastructure

```bash
docker compose up -d
```

This brings up:

- **Postgres 16** on `localhost:5432`
  database `goff_mgmt`, user `goff`, password `goff`
- **Keycloak 25** on `http://localhost:8081`
  pre-imports the `goff` realm with:
  - client `goff-mgmt` (secret `goff-mgmt-secret`)
  - redirect URI `http://localhost:8080/auth/callback`
  - user `admin@local` / password `admin`

Wait until `docker compose ps` shows both healthy.

---

## 4. Configure

Copy and edit:

```bash
cp .env.example .env
```

All variables use the prefix `GOFF_MGMT_`. Snake-cased segments map to dotted
config keys (e.g. `GOFF_MGMT_DB_URL` → `db.url`).

| Variable | Purpose |
|---|---|
| `GOFF_MGMT_SERVER_PORT` | HTTP port (default 8080) |
| `GOFF_MGMT_DB_URL` | Postgres DSN |
| `GOFF_MGMT_OIDC_ISSUER` | OIDC issuer URL |
| `GOFF_MGMT_OIDC_CLIENTID` | OIDC client id |
| `GOFF_MGMT_OIDC_CLIENTSECRET` | OIDC client secret |
| `GOFF_MGMT_OIDC_REDIRECTURL` | OIDC redirect URL (must match the IdP) |
| `GOFF_MGMT_AUTH_JWTSECRET` | HMAC key for session JWT (min 32 chars) |
| `GOFF_MGMT_AUTH_COOKIESECURE` | `true` in prod (HTTPS), `false` for local HTTP |
| `GOFF_MGMT_AUTH_ADMINEMAILS` | Comma-separated emails auto-promoted to super admin on first login |
| `GOFF_MGMT_LOG_LEVEL` | `debug|info|warn|error` |
| `GOFF_MGMT_LOG_FORMAT` | `json` or `console` |

---

## 5. Apply migrations

```bash
./scripts/bootstrap.sh
```

The script waits for Postgres, runs all `goose` migrations from
`db/migrations/`, then (unless `GOFF_MGMT_SEED=0`) applies the optional seed in
`scripts/seed.sql`.

Equivalent manual command:

```bash
goose -dir ./db/migrations postgres "$GOFF_MGMT_DB_URL" up
```

---

## 6. Run the API

From the repo root:

```bash
set -a && source cmd/management/.env && set +a
go run ./cmd/management
```

You should see `starting server addr=:8080`.

If you prefer live reload, install [`air`](https://github.com/air-verse/air) and run `air` from `cmd/management`.

---

## 7. Try it out

1. Open Swagger UI: <http://localhost:8080/swagger/index.html>
   *Generate / refresh the spec first with `swag init -g main.go -o docs` from this directory.*
2. Start an OIDC login: <http://localhost:8080/auth/login>
   You will be redirected to Keycloak (`admin@local` / `admin`).
3. After callback the browser is redirected to `/` with a `goff_mgmt_session`
   cookie. Because `admin@local` is in `GOFF_MGMT_AUTH_ADMINEMAILS`, that user
   is now super-admin.
4. From the same browser:
   ```
   GET  http://localhost:8080/api/v1/auth/me
   POST http://localhost:8080/api/v1/teams       {"name":"platform"}
   GET  http://localhost:8080/api/v1/teams
   ```
5. With `curl`, keep the cookie jar:
   ```bash
   curl -c cookies.txt -b cookies.txt -L http://localhost:8080/auth/login
   curl -b cookies.txt http://localhost:8080/api/v1/auth/me
   ```

---

## 8. Common dev tasks

| Goal | Command |
|---|---|
| Regenerate Swagger | `swag init -g main.go -o docs` |
| Run unit tests | `go test ./...` |
| Run integration tests (require Docker for testcontainers) | `go test -tags=integration ./...` |
| Reset the DB | `goose -dir ./db/migrations postgres "$GOFF_MGMT_DB_URL" reset` |
| Tear down infra | `docker compose down -v` |
| Lint | `golangci-lint run ./...` (from repo root preferred) |

---

## 9. Connect to the database with DBeaver

The local Postgres container is reachable from the host on `localhost:5432`.

1. Open DBeaver → **Database → New Database Connection** → **PostgreSQL**.
2. Fill in:

   | Field | Value |
   |---|---|
   | Host | `localhost` |
   | Port | `5432` |
   | Database | `goff_mgmt` |
   | Username | `goff` |
   | Password | `goff` |
   | Save password | yes (local dev only) |

3. **Driver properties** tab: leave defaults. SSL is disabled in dev (matches the `sslmode=disable` in `GOFF_MGMT_DB_URL`).
4. Click **Test Connection** → **Finish**.

Equivalent JDBC URL (paste into the *URL* field if you prefer):

```
jdbc:postgresql://localhost:5432/goff_mgmt
```

Quick sanity checks once connected:

```sql
SELECT table_name FROM information_schema.tables WHERE table_schema = 'public';
SELECT * FROM goose_db_version ORDER BY id DESC;   -- migration history
SELECT * FROM users;
SELECT * FROM teams;
SELECT * FROM audit_log ORDER BY occurred_at DESC LIMIT 50;
```

If the connection fails, confirm the container is up: `docker compose ps` and
`docker compose logs postgres`.

---

## 10. Troubleshooting

| Symptom | Likely cause / fix |
|---|---|
| `auth.jwtSecret must be at least 32 chars` on startup | Set `GOFF_MGMT_AUTH_JWTSECRET` to a longer string |
| OIDC: `invalid redirect_uri` | The URL in `.env` must match the Keycloak client exactly |
| `Get "https://proxy.golang.org/...": dial tcp` | No network; configure `GOPROXY` or run from a host with internet access |
| Migrations refuse to run | Confirm `GOFF_MGMT_DB_URL` is reachable; `psql` to verify credentials |
| Swagger 404 | Run `swag init` then restart the server |
| Port 5432 / 8080 / 8081 already in use | Edit the relevant ports in `docker-compose.yml` and `.env` |

---

## 11. Code layout

```
cmd/management/
├── config/      koanf-based loader + validator
├── db/migrations/  goose SQL migrations
├── model/       entities and request/response DTOs
├── repository/  pgx repositories (Postgres only)
├── service/     business logic; OIDC + JWT, validation against modules/core/dto
├── api/         Echo server, middleware (auth/RBAC/log), route registration
│   └── middleware/
├── handler/     Echo handlers; thin layer over services
├── scripts/     bootstrap.sh, seed.sql, keycloak realm import
├── docs/        generated Swagger (do not edit by hand)
├── docker-compose.yml
└── .env.example
```

### Conventions
- Layered: handler → service → repository → DB. No direct repo access from handlers.
- All response bodies wrapped in `model.APIResponse` / `model.PaginatedResponse[T]`.
- Validation errors returned as `model.ValidationErrors`; surfaced as 400 by handlers.
- Table-driven tests with `testify`.
- Commits use semantic prefix (`feat:`, `fix:`, `chore:`, etc.). PR titles too.
- Run `make lint` from the repo root before pushing.
