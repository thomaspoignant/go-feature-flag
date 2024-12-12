package postgresqlretriever

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
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
	dbPool *pgxpool.Pool
	status string
	logger *fflog.FFLogger
}

func (r *Retriever) Init(ctx context.Context, logger *fflog.FFLogger) error {
	r.logger = logger
	if r.dbPool == nil {
		r.status = retriever.RetrieverNotReady

		poolConfig, err := pgxpool.ParseConfig(r.URI)
		if err != nil {
			// This will not be a connection error, but a DSN parse error or
			// another initialization error.
			r.status = retriever.RetrieverError
			return err
		}

		pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
		if err != nil {
			r.status = retriever.RetrieverError
			return err
		}

		// Test the database connection
		if err := pool.Ping(ctx); err != nil {
			r.status = retriever.RetrieverError
			return err
		}

		r.dbPool = pool
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
func (r *Retriever) Shutdown(_ context.Context) error {
	r.dbPool.Close()
	return nil
}

// Retrieve Reads flag configuration from postgreSQL and returns it
// If a document does not comply with the specification it will be ignored
func (r *Retriever) Retrieve(ctx context.Context) ([]byte, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s IS NOT NULL", r.Column, r.Table, r.Column)
	rows, err := r.dbPool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ffDocs := make(map[string]interface{})
	for rows.Next() {
		var jsonData []byte

		err := rows.Scan(&jsonData)
		if err != nil {
			return nil, err
		}

		var doc map[string]interface{}
		err = json.Unmarshal(jsonData, &doc)

		if err != nil {
			r.logger.Error("Failed to unmarshal row:", err)
			continue
		}

		if val, ok := doc["flag"]; ok {
			delete(doc, "flag")
			if str, ok := val.(string); ok {
				ffDocs[str] = doc
			} else {
				r.logger.Error("Flag key does not have a string value")
			}
		} else {
			r.logger.Error("No 'flag' entry found")
		}
	}

	flags, err := json.Marshal(ffDocs)

	if err != nil {
		return nil, err
	}

	return flags, nil
}
