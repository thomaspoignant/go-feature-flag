package flaghistory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thomaspoignant/go-feature-flag/retriever/postgresqlretriever"
)

const (
	// DefaultTable is the table used when none is configured.
	DefaultTable = "flag_history"
	// DefaultLimit is the page size used when none is asked.
	DefaultLimit = 50
	// MaxLimit is the biggest page size allowed.
	MaxLimit = 500
)

// Store persists the history.
type Store interface {
	// Migrate creates the schema if it does not exist yet, it is idempotent.
	Migrate(ctx context.Context) error
	// Record inserts the entry, unless the latest known state of the flag is already the same.
	// It returns true when a new row has been inserted.
	Record(ctx context.Context, e Entry) (bool, error)
	// Latest returns the latest entry of a flag, nil if the flag has no history.
	Latest(ctx context.Context, flagset, flagKey string) (*Entry, error)
	// LatestPerFlag returns the latest entry of every flag of a flagset.
	LatestPerFlag(ctx context.Context, flagset string) (map[string]Entry, error)
	// List returns a page of the history, most recent first.
	List(ctx context.Context, f Filter) (Page, error)
	// Purge deletes the entries older than the date and returns the number of deleted rows.
	Purge(ctx context.Context, before time.Time) (int64, error)
	// Ping checks that the database is reachable.
	Ping(ctx context.Context) error
	// Close releases the connection.
	Close(ctx context.Context)
}

// PostgresStore is the PostgreSQL implementation of the Store.
// It shares the connection pool with the PostgreSQL retriever.
type PostgresStore struct {
	uri   string
	table string
	pool  *pgxpool.Pool
}

// NewPostgresStore connects to the database. An empty table means DefaultTable.
func NewPostgresStore(ctx context.Context, uri, table string) (*PostgresStore, error) {
	if table == "" {
		table = DefaultTable
	}
	pool, err := postgresqlretriever.GetPool(ctx, uri)
	if err != nil {
		return nil, fmt.Errorf("impossible to connect to the flag history database: %w", err)
	}
	return &PostgresStore{uri: uri, table: table, pool: pool}, nil
}

func (s *PostgresStore) tableName() string {
	return pgx.Identifier{s.table}.Sanitize()
}

func (s *PostgresStore) indexName(suffix string) string {
	return pgx.Identifier{s.table + "_" + suffix}.Sanitize()
}

func (s *PostgresStore) Migrate(ctx context.Context) error {
	statements := []string{
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			id           BIGSERIAL PRIMARY KEY,
			flagset      TEXT NOT NULL,
			flag_key     TEXT NOT NULL,
			action       TEXT NOT NULL,
			diff         JSONB NOT NULL,
			after_hash   TEXT NOT NULL,
			author       TEXT NOT NULL,
			author_email TEXT,
			change_ref   TEXT,
			source       TEXT,
			detected_at  TIMESTAMPTZ NOT NULL DEFAULT now()
		)`, s.tableName()),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS %s ON %s (flagset, flag_key, id DESC)`,
			s.indexName("flag_idx"), s.tableName()),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS %s ON %s (detected_at)`,
			s.indexName("detected_idx"), s.tableName()),
	}
	for _, stmt := range statements {
		if _, err := s.pool.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("impossible to migrate the flag history table: %w", err)
		}
	}
	return nil
}

func (s *PostgresStore) Record(ctx context.Context, e Entry) (bool, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Every relay proxy instance sees the same change: the lock serializes them per flag,
	// and the second one finds the row of the first one.
	lockKey := e.Flagset + "/" + e.FlagKey
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, lockKey); err != nil {
		return false, err
	}

	var lastAction, lastHash string
	err = tx.QueryRow(ctx, fmt.Sprintf(
		`SELECT action, after_hash FROM %s WHERE flagset = $1 AND flag_key = $2 ORDER BY id DESC LIMIT 1`,
		s.tableName()), e.Flagset, e.FlagKey).Scan(&lastAction, &lastHash)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
	case err != nil:
		return false, err
	default:
		lastAlive := lastAction != string(ActionDeleted)
		if !lastAlive && e.Action == ActionDeleted {
			return false, nil
		}
		if lastAlive && e.Action != ActionDeleted {
			if lastHash == e.AfterHash {
				return false, nil
			}
			if e.Action == ActionAdded {
				// the flag was already known: it is the first load after a restart, so the
				// flag has been modified while we were not watching it.
				e.Action = ActionUpdated
				e.Diff = Diff{Reconciled: true}
			}
		}
	}
	diff, err := json.Marshal(e.Diff)
	if err != nil {
		return false, fmt.Errorf("impossible to marshal the diff: %w", err)
	}

	_, err = tx.Exec(ctx, fmt.Sprintf(
		`INSERT INTO %s (flagset, flag_key, action, diff, after_hash, author, author_email, change_ref, source)
		 VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''), NULLIF($8, ''), NULLIF($9, ''))`, s.tableName()),
		e.Flagset, e.FlagKey, string(e.Action), diff, e.AfterHash, e.Author, e.AuthorEmail, e.ChangeRef, e.Source)
	if err != nil {
		return false, err
	}
	return true, tx.Commit(ctx)
}

const selectColumns = `id, flagset, flag_key, action, diff, after_hash, author,
	COALESCE(author_email, ''), COALESCE(change_ref, ''), COALESCE(source, ''), detected_at`

func scanEntry(row pgx.Row) (Entry, error) {
	var e Entry
	var action string
	var diff []byte
	err := row.Scan(&e.ID, &e.Flagset, &e.FlagKey, &action, &diff, &e.AfterHash, &e.Author,
		&e.AuthorEmail, &e.ChangeRef, &e.Source, &e.DetectedAt)
	if err != nil {
		return Entry{}, err
	}
	e.Action = Action(action)
	if err := json.Unmarshal(diff, &e.Diff); err != nil {
		return Entry{}, fmt.Errorf("invalid diff stored for the entry %d: %w", e.ID, err)
	}
	return e, nil
}

func (s *PostgresStore) Latest(ctx context.Context, flagset, flagKey string) (*Entry, error) {
	row := s.pool.QueryRow(ctx, fmt.Sprintf(
		`SELECT %s FROM %s WHERE flagset = $1 AND flag_key = $2 ORDER BY id DESC LIMIT 1`,
		selectColumns, s.tableName()), flagset, flagKey)
	e, err := scanEntry(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (s *PostgresStore) LatestPerFlag(ctx context.Context, flagset string) (map[string]Entry, error) {
	rows, err := s.pool.Query(ctx, fmt.Sprintf(
		`SELECT DISTINCT ON (flag_key) %s FROM %s WHERE flagset = $1 ORDER BY flag_key, id DESC`,
		selectColumns, s.tableName()), flagset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	latest := map[string]Entry{}
	for rows.Next() {
		e, err := scanEntry(rows)
		if err != nil {
			return nil, err
		}
		latest[e.FlagKey] = e
	}
	return latest, rows.Err()
}

func (s *PostgresStore) List(ctx context.Context, f Filter) (Page, error) {
	limit := f.Limit
	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	where := make([]string, 0, 7)
	args := make([]any, 0, 8)
	add := func(cond string, v any) {
		args = append(args, v)
		where = append(where, fmt.Sprintf(cond, len(args)))
	}
	if f.Flagset != "" {
		add("flagset = $%d", f.Flagset)
	}
	if f.FlagKey != "" {
		add("flag_key = $%d", f.FlagKey)
	}
	if f.Action != "" {
		add("action = $%d", string(f.Action))
	}
	if f.Author != "" {
		add("author = $%d", f.Author)
	}
	if !f.From.IsZero() {
		add("detected_at >= $%d", f.From)
	}
	if !f.To.IsZero() {
		add("detected_at <= $%d", f.To)
	}
	if f.Cursor > 0 {
		add("id < $%d", f.Cursor)
	}
	query := fmt.Sprintf(`SELECT %s FROM %s`, selectColumns, s.tableName())
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	// we ask one more row to know if there is a next page.
	args = append(args, limit+1)
	query += fmt.Sprintf(" ORDER BY id DESC LIMIT $%d", len(args))

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return Page{}, err
	}
	defer rows.Close()
	page := Page{Entries: make([]Entry, 0, limit)}
	for rows.Next() {
		e, err := scanEntry(rows)
		if err != nil {
			return Page{}, err
		}
		page.Entries = append(page.Entries, e)
	}
	if err := rows.Err(); err != nil {
		return Page{}, err
	}
	if len(page.Entries) > limit {
		page.Entries = page.Entries[:limit]
		page.NextCursor = page.Entries[limit-1].ID
	}
	return page, nil
}

// Purge keeps the latest entry of every flag still alive: it is the baseline used to detect
// the changes made while the relay proxy was stopped.
func (s *PostgresStore) Purge(ctx context.Context, before time.Time) (int64, error) {
	tag, err := s.pool.Exec(ctx, fmt.Sprintf(
		`DELETE FROM %[1]s WHERE detected_at < $1 AND id NOT IN (
			SELECT id FROM (
				SELECT DISTINCT ON (flagset, flag_key) id, action FROM %[1]s
				ORDER BY flagset, flag_key, id DESC
			) latest WHERE action <> 'deleted'
		)`, s.tableName()), before)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (s *PostgresStore) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

func (s *PostgresStore) Close(ctx context.Context) {
	postgresqlretriever.ReleasePool(ctx, s.uri)
}
