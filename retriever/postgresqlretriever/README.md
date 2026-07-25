# PostgreSQL Feature Flag Retriever

This retriever is used to retrieve feature flag configurations from a PostgreSQL database.

## Installation

```bash
go get github.com/thomaspoignant/go-feature-flag/retriever/postgresqlretriever
```

## Usage

### Database Schema

The retriever requires a table with these **minimum columns**:

- `flag_name` (VARCHAR): The name of the feature flag
- `flagset` (VARCHAR): The flagset/namespace for the flag (typically "default")
- `config` (JSONB): The feature flag configuration as JSON

> **Note**: These are the default column names. You can use different column names in your table and map them using the `Columns` field in the retriever configuration.

#### Example Schema

```sql
-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create the go_feature_flag table
CREATE TABLE IF NOT EXISTS go_feature_flag (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    flag_name VARCHAR(255) NOT NULL,
    flagset VARCHAR(255) NOT NULL,
    config JSONB NOT NULL
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_go_feature_flag_flagset ON go_feature_flag(flagset);
CREATE INDEX IF NOT EXISTS idx_go_feature_flag_flag_name ON go_feature_flag(flag_name);
CREATE INDEX IF NOT EXISTS idx_go_feature_flag_flagset_flag_name ON go_feature_flag(flagset, flag_name);

-- Add a unique constraint to prevent duplicate flags in the same flagset
ALTER TABLE go_feature_flag ADD CONSTRAINT unique_flag_per_flagset UNIQUE (flag_name, flagset);
```

### Configuration

Create a `Retriever` struct with the following fields:

- `URI`: PostgreSQL connection string (required)
- `Table`: Name of the table containing feature flags (required)
- `Columns`: (Optional) Custom column name mapping

#### Basic Configuration

```go
retriever := &postgresqlretriever.Retriever{
    URI:   "postgres://user:password@localhost:5432/dbname",
    Table: "go_feature_flag",
}
```

#### Custom Column Names

If your table uses different column names than the defaults (`flag_name`, `flagset`, `config`), you can customize the mapping using the `Columns` field:

```go
retriever := &postgresqlretriever.Retriever{
    URI:   "postgres://user:password@localhost:5432/dbname",
    Table: "my_feature_flags",
    Columns: map[string]string{
        "flag_name": "name",      // Your column name for flag names
        "flagset":   "namespace", // Your column name for flagsets
        "config":    "settings",  // Your column name for config JSON
    },
}
```

## Key Features

- **Flexible Schema**: Supports custom table and column names
- **Safe SQL Queries**: Uses PostgreSQL identifier sanitization to prevent SQL injection
- **Connection Management**: Automatic connection initialization and cleanup
- **Error Handling**: Comprehensive error handling with detailed logging
- **Performance Optimized**: Efficient querying with proper indexing support
- **Write support**: The only retriever that backs the relay proxy's flag-management API
  (create/update/delete), because it can guarantee optimistic concurrency through real
  database transactions (see below).

## Flag Management API (relay proxy)

The relay proxy exposes a write API (`POST/GET/PUT/PATCH/DELETE /v1/flags[/{flag_key}]`,
`PATCH /v1/flags/{flag_key}/state`) that only works when a flagset's retriever is this
PostgreSQL retriever. Every other retriever kind (file, S3, GitHub, ...) has no
compare-and-swap primitive and cannot be used for these endpoints.

Requirements for a flagset to be writable:

- It must have **exactly one** retriever configured, and it must be this PostgreSQL retriever.
- The retriever must have an **explicit, non-empty `flagset` configured** (the `flagset`
  column is `NOT NULL`, so there is no single unambiguous row to write to when no flagset is
  set — a flagset without an explicit name always responds `403 FLAG_CONFIG` on writes, even
  though it can still be read).
- The table must keep the `UNIQUE (flag_name, flagset)` constraint from the schema above:
  writes rely on it for safe upserts.

Concurrency: writes take place inside a transaction that locks the target row with
`SELECT ... FOR UPDATE` before comparing the caller-supplied `If-Match` ETag against the
current one, so two concurrent writers racing the same flag can never both succeed silently —
the loser gets `412 Precondition Failed`. The ETag is a strong hash computed over the
canonical JSON of the flag definition, not over the raw JSONB bytes, so it stays stable
regardless of Postgres's own JSON formatting.

Every successful write triggers an immediate cache reload, so evaluation reflects the change
without waiting for the next polling interval.
