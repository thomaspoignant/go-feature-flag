package postgresqlretriever

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/utils"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

var defaultColumns = map[string]string{
	"flag_name": "flag_name",
	"flagset":   "flagset",
	"config":    "config",
}

type Retriever struct {
	URI     string
	Table   string
	Columns map[string]string

	logger  *fflog.FFLogger
	status  retriever.Status
	columns map[string]string
	pool    *pgxpool.Pool
	flagset *string
}

func (r *Retriever) Init(ctx context.Context, logger *fflog.FFLogger, flagset *string) error {
	r.status = retriever.RetrieverNotReady
	r.logger = logger
	if r.logger == nil {
		r.logger = &fflog.FFLogger{}
	}

	r.columns = r.getColumnNames()
	r.flagset = flagset

	if r.pool == nil {
		r.logger.Info("Initializing PostgreSQL retriever")
		r.logger.Debug("Using columns", "columns", r.columns)

		pool, err := GetPool(ctx, r.URI)
		if err != nil {
			r.status = retriever.RetrieverError
			return err
		}

		r.pool = pool
	}

	r.status = retriever.RetrieverReady
	return nil
}

// Status returns the current status of the retriever
func (r *Retriever) Status() retriever.Status {
	if r == nil || r.status == "" {
		return retriever.RetrieverNotReady
	}
	return r.status
}

// Shutdown closes the database connection.
func (r *Retriever) Shutdown(ctx context.Context) error {
	ReleasePool()
	return nil 
}

// Retrieve fetches flag configuration from PostgreSQL.
func (r *Retriever) Retrieve(ctx context.Context) ([]byte, error) {
	if r.pool == nil {
		return nil, fmt.Errorf("database connection pool is not initialized")
	}

	// Build the query using the configured table and column names
	query := r.buildQuery()

	// Build the arguments for the query
	args := []any{}
	if r.getFlagset() != "" {
		// If a flagset is defined, it becomes the first ($1) argument.
		args = []any{r.getFlagset()}
	}

	r.logger.Debug("Executing PostgreSQL query", slog.String("query", query), slog.Any("args", args))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Map to store flag configurations with flag_name as key
	flagConfigs := make(map[string]any)

	for rows.Next() {
		var flagName string
		var configData []byte

		if err := rows.Scan(&flagName, &configData); err != nil {
			r.logger.Error("Failed to scan row", "error", err)
			continue
		}

		// Parse the JSON config data into a map
		var config map[string]any
		if err := json.Unmarshal(configData, &config); err != nil {
			r.logger.Error("Failed to unmarshal config data", "error", err, "flagName", flagName)
			continue
		}
		flagConfigs[flagName] = config
	}

	// Check for any errors that occurred during row iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	// If no flags found, return empty JSON object
	if len(flagConfigs) == 0 {
		return []byte("{}"), nil
	}

	// Marshal the flag configurations to JSON
	result, err := json.Marshal(flagConfigs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal flag configurations: %w", err)
	}

	r.logger.Debug("Retrieved flags from PostgreSQL", "count", len(flagConfigs))
	return result, nil
}

// buildQuery constructs the SQL query and arguments based on whether a flagset is specified.
// It uses pgx.Identifier to safely quote identifiers and prevent SQL injection.
func (r *Retriever) buildQuery() string {
	var query string

	flagNameCol := pgx.Identifier{r.columns["flag_name"]}.Sanitize()
	configCol := pgx.Identifier{r.columns["config"]}.Sanitize()
	tableCol := pgx.Identifier{r.Table}.Sanitize()
	flagsetCol := pgx.Identifier{r.columns["flagset"]}.Sanitize()

	if r.getFlagset() != "" {
		query = fmt.Sprintf("SELECT %s, %s FROM %s WHERE %s = $1 ORDER BY %s",
			flagNameCol, configCol, tableCol, flagsetCol, flagNameCol)
	} else {
		query = fmt.Sprintf("SELECT %s, %s FROM %s ORDER BY %s",
			flagNameCol, configCol, tableCol, flagNameCol)
	}

	return query
}

// getColumnNames returns the column names to use for database queries.
// If r.Columns is provided, it merges those values with defaultColumns,
// using defaults for any missing entries.
func (r *Retriever) getColumnNames() map[string]string {
	// Start with a copy of defaultColumns
	columns := make(map[string]string)
	for k, v := range defaultColumns {
		columns[k] = v
	}

	// Override with any provided custom columns
	if r.Columns != nil {
		for k, v := range r.Columns {
			columns[k] = v
		}
	}

	return columns
}

func (r *Retriever) getFlagset() string {
	if r.flagset != nil && *r.flagset != "" && *r.flagset != utils.DefaultFlagSetName {
		return *r.flagset
	}
	return ""
}
