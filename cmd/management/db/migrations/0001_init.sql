-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "citext";

CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email           CITEXT NOT NULL UNIQUE,
    name            TEXT NOT NULL DEFAULT '',
    oidc_sub        TEXT UNIQUE,
    is_super_admin  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login_at   TIMESTAMPTZ
);

CREATE TABLE teams (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE team_members (
    team_id     UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role        TEXT NOT NULL CHECK (role IN ('admin','editor','viewer')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (team_id, user_id)
);
CREATE INDEX idx_team_members_user ON team_members(user_id);

CREATE TABLE flagsets (
    id                   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id              UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    name                 TEXT NOT NULL,
    description          TEXT NOT NULL DEFAULT '',
    api_key_hashes       TEXT[] NOT NULL DEFAULT '{}',
    polling_interval_ms  INTEGER,
    file_format          TEXT,
    retrievers           JSONB,
    exporters            JSONB,
    notifiers            JSONB,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (team_id, name)
);

CREATE TABLE flags (
    id                 UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    flagset_id         UUID NOT NULL REFERENCES flagsets(id) ON DELETE CASCADE,
    name               TEXT NOT NULL,
    current_version_id UUID,
    disabled           BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at         TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (flagset_id, name)
);

CREATE TABLE flag_versions (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    flag_id        UUID NOT NULL REFERENCES flags(id) ON DELETE CASCADE,
    version_number INTEGER NOT NULL,
    payload        JSONB NOT NULL,
    comment        TEXT NOT NULL DEFAULT '',
    created_by     UUID REFERENCES users(id),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (flag_id, version_number)
);
CREATE INDEX idx_flag_versions_flag_desc ON flag_versions(flag_id, version_number DESC);

ALTER TABLE flags
    ADD CONSTRAINT fk_flags_current_version
    FOREIGN KEY (current_version_id) REFERENCES flag_versions(id) ON DELETE SET NULL;

CREATE TABLE audit_log (
    id              BIGSERIAL PRIMARY KEY,
    occurred_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    actor_user_id   UUID REFERENCES users(id),
    team_id         UUID REFERENCES teams(id) ON DELETE SET NULL,
    flagset_id      UUID REFERENCES flagsets(id) ON DELETE SET NULL,
    flag_id         UUID REFERENCES flags(id) ON DELETE SET NULL,
    action          TEXT NOT NULL,
    target_type     TEXT NOT NULL,
    target_id       TEXT NOT NULL,
    before          JSONB,
    after           JSONB,
    metadata        JSONB
);
CREATE INDEX idx_audit_log_occurred_desc ON audit_log(occurred_at DESC);
CREATE INDEX idx_audit_log_team_occurred ON audit_log(team_id, occurred_at DESC);
CREATE INDEX idx_audit_log_flagset      ON audit_log(flagset_id);
CREATE INDEX idx_audit_log_flag         ON audit_log(flag_id);
CREATE INDEX idx_audit_log_actor        ON audit_log(actor_user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS audit_log;
ALTER TABLE IF EXISTS flags DROP CONSTRAINT IF EXISTS fk_flags_current_version;
DROP TABLE IF EXISTS flag_versions;
DROP TABLE IF EXISTS flags;
DROP TABLE IF EXISTS flagsets;
DROP TABLE IF EXISTS team_members;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
