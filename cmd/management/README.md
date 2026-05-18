# GO Feature Flag — Management API

Source-of-truth REST API for managing GO Feature Flag flagsets and flags,
with team ownership, OIDC authentication, audit log, and per-flag version
history with rollback.

Hierarchy:

```
Team (members with roles: admin | editor | viewer)
  └── Flagset (apiKeys, retrievers/exporters/notifiers config)
        └── Flag (versions, current version, audit trail)
```

See [`CONTRIBUTING.md`](./CONTRIBUTING.md) for local-environment setup
(`docker compose up -d`, migrations, OIDC config, running the server).

## Scope

- CRUD for teams, flagsets, flags
- OIDC login (generic, config-driven)
- Per-flag versioning with rollback
- Full audit log with filters
- Postgres-backed (pgx v5)

## Out of scope for the first iteration

- Frontend UI
- Relay-proxy integration (this API does not yet expose flags to the relay proxy; see plan)
- Multi-team flagset sharing
- Webhooks / change notifications dispatched by this service

## Module boundaries

This service is its own Go module (`cmd/management/go.mod`). Within the
GO Feature Flag monorepo, it depends only on `modules/core` — never on
`cmdhelpers`, `cmd/relayproxy`, `retriever/*`, `exporter/*`, `internal/*`,
or the root `ffclient` package.
