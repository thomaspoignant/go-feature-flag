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
