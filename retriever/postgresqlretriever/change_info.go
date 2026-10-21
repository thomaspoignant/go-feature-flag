package postgresqlretriever

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/thomaspoignant/go-feature-flag/notifier"
)

// LastChange returns who made the last modification of the flags.
// It is available only when the "updated_by" and "updated_at" columns are configured in Columns,
// it returns nil otherwise.
func (r *Retriever) LastChange(ctx context.Context) (*notifier.ChangeInfo, error) {
	updatedBy, updatedAt := r.Columns["updated_by"], r.Columns["updated_at"]
	if updatedBy == "" || updatedAt == "" {
		return nil, nil
	}
	if r.pool == nil {
		return nil, fmt.Errorf("database connection pool is not initialized")
	}

	query := fmt.Sprintf("SELECT %s, %s FROM %s",
		pgx.Identifier{updatedBy}.Sanitize(), pgx.Identifier{updatedAt}.Sanitize(),
		pgx.Identifier{r.Table}.Sanitize())
	args := []any{}
	if r.getFlagset() != "" {
		query += fmt.Sprintf(" WHERE %s = $1", pgx.Identifier{r.columns["flagset"]}.Sanitize())
		args = append(args, r.getFlagset())
	}
	query += fmt.Sprintf(" ORDER BY %s DESC LIMIT 1", pgx.Identifier{updatedAt}.Sanitize())

	var author *string
	var changedAt *time.Time
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&author, &changedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find the last change: %w", err)
	}
	if author == nil || *author == "" {
		return nil, nil
	}
	info := &notifier.ChangeInfo{Author: *author}
	if changedAt != nil {
		info.ChangedAt = *changedAt
	}
	return info, nil
}
