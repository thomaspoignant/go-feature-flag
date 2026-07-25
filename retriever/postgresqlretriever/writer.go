package postgresqlretriever

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/thomaspoignant/go-feature-flag/retriever"
)

// GetFlag returns the JSON definition bytes and the current ETag of a single flag.
// Returns retriever.ErrFlagNotFound if the flag doesn't exist.
func (r *Retriever) GetFlag(threadContext context.Context, flagKey string) ([]byte, string, error) {
	if r.pool == nil {
		return nil, "", fmt.Errorf("database connection pool is not initialized")
	}

	flagNameCol := pgx.Identifier{r.columns["flag_name"]}.Sanitize()
	configCol := pgx.Identifier{r.columns["config"]}.Sanitize()
	tableCol := pgx.Identifier{r.Table}.Sanitize()
	flagsetCol := pgx.Identifier{r.columns["flagset"]}.Sanitize()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1 AND %s = $2",
		configCol, tableCol, flagNameCol, flagsetCol)
	row := r.pool.QueryRow(threadContext, query, flagKey, r.getFlagset())
	var flagDefinition []byte
	if scanError := row.Scan(&flagDefinition); scanError != nil {
		if errors.Is(scanError, pgx.ErrNoRows) {
			return nil, "", retriever.ErrFlagNotFound
		}
		return nil, "", fmt.Errorf("failed to execute query: %w", scanError)
	}

	etag, computeETagErr := computeETag(flagDefinition)
	if computeETagErr != nil {
		return nil, "", computeETagErr
	}
	return flagDefinition, etag, nil
}

// CreateFlag creates a new flag. Returns retriever.ErrFlagAlreadyExists if the key already
// exists, and retriever.ErrFlagsetNotConfigured if the retriever has no explicit flagset.
func (r *Retriever) CreateFlag(threadContext context.Context, flagKey string, definition []byte) (string, error) {
	if validateWritePrerequisitesErr := r.validateWritePrerequisites(); validateWritePrerequisitesErr != nil {
		return "", validateWritePrerequisitesErr
	}

	transaction, rollbackTransaction, beginWriteTransactionErr := r.beginWriteTransactionWithRollback(threadContext)
	if beginWriteTransactionErr != nil {
		return "", beginWriteTransactionErr
	}
	defer rollbackTransaction()

	_, flagAlreadyExists, selectFlagErr := r.selectForUpdate(threadContext, transaction, flagKey)
	if selectFlagErr != nil {
		return "", selectFlagErr
	}
	if flagAlreadyExists {
		return "", retriever.ErrFlagAlreadyExists
	}

	if insertFlagError := r.insertFlag(threadContext, transaction, flagKey, definition); insertFlagError != nil {
		return "", insertFlagError
	}
	if commitTransactionErr := transaction.Commit(threadContext); commitTransactionErr != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", commitTransactionErr)
	}
	return computeETag(definition)
}

// UpsertFlag replaces the flag definition, creating it if it doesn't exist yet (created=true).
// When ifMatch is non-nil and the flag already exists, its current ETag must match or
// retriever.ErrETagMismatch is returned. ifMatch is ignored when the flag doesn't exist yet.
func (r *Retriever) UpsertFlag(
	threadContext context.Context, flagKey string, definition []byte, ifMatch *string,
) (bool, string, error) {
	if validateWritePrerequisitesErr := r.validateWritePrerequisites(); validateWritePrerequisitesErr != nil {
		return false, "", validateWritePrerequisitesErr
	}

	transaction, rollbackTransaction, beginWriteTransactionErr := r.beginWriteTransactionWithRollback(threadContext)
	if beginWriteTransactionErr != nil {
		return false, "", beginWriteTransactionErr
	}
	defer rollbackTransaction()

	currentDefinition, flagFoundInDatabase, selectFlagErr := r.selectForUpdate(threadContext, transaction, flagKey)
	if selectFlagErr != nil {
		return false, "", selectFlagErr
	}

	if flagFoundInDatabase {
		if etagMathError := r.validateIfMatchETag(currentDefinition, ifMatch); etagMathError != nil {
			return false, "", etagMathError
		}
		if updateFlagErr := r.updateFlag(threadContext, transaction, flagKey, definition); updateFlagErr != nil {
			return false, "", updateFlagErr
		}
	} else {
		if insertFlagErr := r.insertFlag(threadContext, transaction, flagKey, definition); insertFlagErr != nil {
			return false, "", insertFlagErr
		}
	}

	if commitTransactionErr := transaction.Commit(threadContext); commitTransactionErr != nil {
		return false, "", fmt.Errorf("failed to commit transaction: %w", commitTransactionErr)
	}
	etag, computeETagErr := computeETag(definition)
	if computeETagErr != nil {
		return false, "", computeETagErr
	}
	return !flagFoundInDatabase, etag, nil
}

// DeleteFlag removes a flag. Returns retriever.ErrFlagNotFound if it doesn't exist. When
// ifMatch is non-nil, it must match the flag's current ETag or retriever.ErrETagMismatch
// is returned.
func (r *Retriever) DeleteFlag(threadContext context.Context, flagKey string, ifMatch *string) error {
	if validateWritePrerequisitesErr := r.validateWritePrerequisites(); validateWritePrerequisitesErr != nil {
		return validateWritePrerequisitesErr
	}

	transaction, rollbackTransaction, beginWriteTransactionErr := r.beginWriteTransactionWithRollback(threadContext)
	if beginWriteTransactionErr != nil {
		return beginWriteTransactionErr
	}
	defer rollbackTransaction()

	currentDefinition, flagAlreadyExists, selectFlagErr := r.selectForUpdate(threadContext, transaction, flagKey)
	if selectFlagErr != nil {
		return selectFlagErr
	}
	if !flagAlreadyExists {
		return retriever.ErrFlagNotFound
	}
	if ifMatch != nil {
		currentETag, computeETagErr := computeETag(currentDefinition)
		if computeETagErr != nil {
			return computeETagErr
		}
		if currentETag != *ifMatch {
			return retriever.ErrETagMismatch
		}
	}

	if deleteFlagErr := r.deleteFlagRow(threadContext, transaction, flagKey); deleteFlagErr != nil {
		return deleteFlagErr
	}
	return transaction.Commit(threadContext)
}

// Source identifies this retriever instance for the flag-management API's FlagResponse.source.
func (r *Retriever) Source() string {
	return "postgresql:" + r.Table
}

// selectForUpdate locks (SELECT ... FOR UPDATE) and returns the current config of a flag row
// within an in-progress transaction. found is false (with a nil error) when the row doesn't
// exist yet. Locking here is what makes the optimistic-concurrency check race-free across
// concurrent relay-proxy instances hitting the same database.
func (r *Retriever) selectForUpdate(threadContext context.Context, transaction pgx.Tx, flagKey string) ([]byte, bool, error) {
	configCol := pgx.Identifier{r.columns["config"]}.Sanitize()
	tableCol := pgx.Identifier{r.Table}.Sanitize()
	flagNameCol := pgx.Identifier{r.columns["flag_name"]}.Sanitize()
	flagsetCol := pgx.Identifier{r.columns["flagset"]}.Sanitize()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1 AND %s = $2 FOR UPDATE",
		configCol, tableCol, flagNameCol, flagsetCol)
	row := transaction.QueryRow(threadContext, query, flagKey, r.getFlagset())
	var flagDefinition []byte
	if scanErr := row.Scan(&flagDefinition); scanErr != nil {
		if errors.Is(scanErr, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("failed to execute query: %w", scanErr)
	}
	return flagDefinition, true, nil
}

func (r *Retriever) insertFlag(threadContext context.Context, transaction pgx.Tx, flagKey string, definition []byte) error {
	tableCol := pgx.Identifier{r.Table}.Sanitize()
	flagNameCol := pgx.Identifier{r.columns["flag_name"]}.Sanitize()
	flagsetCol := pgx.Identifier{r.columns["flagset"]}.Sanitize()
	configCol := pgx.Identifier{r.columns["config"]}.Sanitize()

	query := fmt.Sprintf("INSERT INTO %s (%s, %s, %s) VALUES ($1, $2, $3)",
		tableCol, flagNameCol, flagsetCol, configCol)
	if _, executeInsertErr := transaction.Exec(threadContext, query, flagKey, r.getFlagset(), definition); executeInsertErr != nil {
		return fmt.Errorf("failed to insert flag: %w", executeInsertErr)
	}
	return nil
}

func (r *Retriever) updateFlag(ctx context.Context, transaction pgx.Tx, flagKey string, definition []byte) error {
	tableCol := pgx.Identifier{r.Table}.Sanitize()
	configCol := pgx.Identifier{r.columns["config"]}.Sanitize()
	flagNameCol := pgx.Identifier{r.columns["flag_name"]}.Sanitize()
	flagsetCol := pgx.Identifier{r.columns["flagset"]}.Sanitize()

	query := fmt.Sprintf("UPDATE %s SET %s = $1 WHERE %s = $2 AND %s = $3",
		tableCol, configCol, flagNameCol, flagsetCol)
	if _, executeUpdateErr := transaction.Exec(ctx, query, definition, flagKey, r.getFlagset()); executeUpdateErr != nil {
		return fmt.Errorf("failed to update flag: %w", executeUpdateErr)
	}
	return nil
}

func (r *Retriever) deleteFlagRow(ctx context.Context, transaction pgx.Tx, flagKey string) error {
	tableCol := pgx.Identifier{r.Table}.Sanitize()
	flagNameCol := pgx.Identifier{r.columns["flag_name"]}.Sanitize()
	flagsetCol := pgx.Identifier{r.columns["flagset"]}.Sanitize()

	query := fmt.Sprintf("DELETE FROM %s WHERE %s = $1 AND %s = $2", tableCol, flagNameCol, flagsetCol)
	if _, executeDeleteErr := transaction.Exec(ctx, query, flagKey, r.getFlagset()); executeDeleteErr != nil {
		return fmt.Errorf("failed to delete flag: %w", executeDeleteErr)
	}
	return nil
}

func (r *Retriever) validateIfMatchETag(
	currentDefinition []byte,
	ifMatch *string,
) error {
	if ifMatch == nil {
		return nil
	}

	currentETag, computeETagErr := computeETag(currentDefinition)
	if computeETagErr != nil {
		return computeETagErr
	}
	if currentETag != *ifMatch {
		return retriever.ErrETagMismatch
	}
	return nil
}

func (r *Retriever) validateWritePrerequisites() error {
	if r.pool == nil {
		return fmt.Errorf("database connection pool is not initialized")
	}
	if r.getFlagset() == "" {
		return retriever.ErrFlagsetNotConfigured
	}
	return nil
}

func (r *Retriever) beginWriteTransaction(threadContext context.Context) (pgx.Tx, error) {
	transaction, beginTransactionErr := r.pool.Begin(threadContext)
	if beginTransactionErr != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", beginTransactionErr)
	}
	return transaction, nil
}

func (r *Retriever) beginWriteTransactionWithRollback(threadContext context.Context) (pgx.Tx, func(), error) {
	transaction, beginWriteTransactionErr := r.beginWriteTransaction(threadContext)
	if beginWriteTransactionErr != nil {
		return nil, nil, beginWriteTransactionErr
	}
	rollbackTransaction := func() {
		_ = transaction.Rollback(threadContext)
	}
	return transaction, rollbackTransaction, nil
}
