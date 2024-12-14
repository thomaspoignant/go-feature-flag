package postgresqlretriever

import (
	"context"
	"encoding/json"
	"fmt"

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
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s IS NOT NULL", r.Column, r.Table, r.Column)
	rows, err := r.conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	mappedFlagDocs := make(map[string]interface{})
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
				mappedFlagDocs[str] = doc
			} else {
				r.logger.Error("Flag key does not have a string value")
			}
		} else {
			r.logger.Warn("No 'flag' entry found")
		}
	}

	flags, err := json.Marshal(mappedFlagDocs)

	if err != nil {
		return nil, err
	}

	return flags, nil
}
