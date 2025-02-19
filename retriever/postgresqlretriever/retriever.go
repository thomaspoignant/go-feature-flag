package postgresqlretriever

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

// Retriever is a configuration struct for a PostgreSQL connection
type Retriever struct {
	// PostgreSQL connection URI
	URI  string
	Type string
	// PostgreSQL table where flag column is stored
	Table string
	// PostgreSQL column where flag definitions are stored
	Column string
	conn   *pgx.Conn
	status string
	logger *fflog.FFLogger
}

func (r *Retriever) Init(ctx context.Context, logger *fflog.FFLogger) error {
	r.logger = logger
	if r.conn == nil {
		r.status = retriever.RetrieverNotReady

		conn, err := pgx.Connect(ctx, r.URI)
		if err != nil {
			r.status = retriever.RetrieverError
			return err
		}

		// Test the database connection
		if err := conn.Ping(ctx); err != nil {
			r.status = retriever.RetrieverError
			return err
		}

		r.conn = conn
		r.status = retriever.RetrieverReady
	}

	return nil
}

// Status returns the current status of the retriever
func (r *Retriever) Status() retriever.Status {
	if r == nil || r.status == "" {
		return retriever.RetrieverNotReady
	}
	return r.status
}

// Shutdown disconnects the retriever from Mongodb instance
func (r *Retriever) Shutdown(ctx context.Context) error {
	r.conn.Close(ctx)
	return nil
}

// Retrieve Reads flag configuration from postgreSQL and returns it
// If a document does not comply with the specification it will be ignored
func (r *Retriever) Retrieve(ctx context.Context) ([]byte, error) {
	var query string
	var err error

	switch r.Type {
	case "json":
		query = fmt.Sprintf("SELECT %s FROM %s WHERE %s IS NOT NULL", r.Column, r.Table, r.Column)
	case "relational":
		query, err = loadSQL("relational_db_query.sql")
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported type: %s", r.Type)
	}

	return executeQuery(ctx, r.conn, query, r.logger)
}

// Executes the given SQL query and returns the result as a slice of bytes.
func executeQuery(ctx context.Context, conn *pgx.Conn, query string, logger *fflog.FFLogger) ([]byte, error) {
	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	mappedFlagDocs := make(map[string]interface{})
	for rows.Next() {
		var data []byte
		// Assuming the data is in JSON format directly retrievable from the database
		if err := rows.Scan(&data); err != nil {
			logger.Error(fmt.Sprintf("Failed to scan row: %v", err))
			continue
		}

		var doc map[string]interface{}
		if err := json.Unmarshal(data, &doc); err != nil {
			logger.Error(fmt.Sprintf("Failed to unmarshal JSON data: %v", err))
			continue
		}

		if val, ok := doc["flag"]; ok {
			delete(doc, "flag")
			if str, ok := val.(string); ok {
				mappedFlagDocs[str] = doc
			} else {
				logger.Error("Flag key does not have a string value")
			}
		} else {
			logger.Warn("No 'flag' entry found")
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	// Only marshal if there is data to marshal
	if len(mappedFlagDocs) == 0 {
		return []byte("{}"), nil
	}

	flags, err := json.Marshal(mappedFlagDocs)
	if err != nil {
		return nil, err
	}

	return flags, nil
}

// Loads an SQL query from an external file.
func loadSQL(filename string) (string, error) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
