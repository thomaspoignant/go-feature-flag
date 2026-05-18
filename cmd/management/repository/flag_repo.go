package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/thomaspoignant/go-feature-flag/cmd/management/model"
)

type FlagRepo struct{ db *DB }

func NewFlagRepo(db *DB) *FlagRepo { return &FlagRepo{db: db} }

func (r *FlagRepo) Create(ctx context.Context, tx pgx.Tx, flagsetID, name string) (*model.Flag, error) {
	var f model.Flag
	err := tx.QueryRow(ctx, `
		INSERT INTO flags (flagset_id, name) VALUES ($1, $2)
		RETURNING id, flagset_id, name, current_version_id, disabled, created_at, updated_at`,
		flagsetID, name,
	).Scan(&f.ID, &f.FlagsetID, &f.Name, &f.CurrentVersionID, &f.Disabled, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *FlagRepo) GetByID(ctx context.Context, id string) (*model.Flag, error) {
	var f model.Flag
	var payload json.RawMessage
	var verNum *int
	err := r.db.Pool.QueryRow(ctx, `
		SELECT f.id, f.flagset_id, f.name, f.current_version_id, f.disabled, f.created_at, f.updated_at,
		       fv.payload, fv.version_number
		FROM flags f
		LEFT JOIN flag_versions fv ON fv.id = f.current_version_id
		WHERE f.id = $1 AND f.deleted_at IS NULL`, id,
	).Scan(&f.ID, &f.FlagsetID, &f.Name, &f.CurrentVersionID, &f.Disabled, &f.CreatedAt, &f.UpdatedAt, &payload, &verNum)
	if err != nil {
		if isNoRows(err) {
			return nil, nil
		}
		return nil, err
	}
	f.Payload = payload
	f.CurrentVersion = verNum
	return &f, nil
}

func (r *FlagRepo) ListByFlagset(ctx context.Context, flagsetID string, filters model.FlagFilters) ([]model.FlagListItem, int, error) {
	qb := NewQueryBuilder()
	if filters.Name != nil && *filters.Name != "" {
		qb.Add("f.name ILIKE $%d", "%"+*filters.Name+"%")
	}
	if filters.Disabled != nil {
		qb.Add("f.disabled = $%d", *filters.Disabled)
	}
	where, args := qb.Build()

	var total int
	countQ := `SELECT COUNT(*) FROM flags f WHERE f.flagset_id = $1 AND f.deleted_at IS NULL` + where
	countArgs := append([]any{flagsetID}, args...)
	if err := r.db.Pool.QueryRow(ctx, countQ, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.PageSize < 1 || filters.PageSize > 200 {
		filters.PageSize = 50
	}
	offset := (filters.Page - 1) * filters.PageSize
	listQ := fmt.Sprintf(`
		SELECT f.id, f.flagset_id, f.name, fv.version_number, f.disabled, f.updated_at
		FROM flags f
		LEFT JOIN flag_versions fv ON fv.id = f.current_version_id
		WHERE f.flagset_id = $1 AND f.deleted_at IS NULL%s
		ORDER BY f.name
		LIMIT $%d OFFSET $%d`, where, len(countArgs)+1, len(countArgs)+2)
	listArgs := append(countArgs, filters.PageSize, offset)

	rows, err := r.db.Pool.Query(ctx, listQ, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []model.FlagListItem
	for rows.Next() {
		var it model.FlagListItem
		if err := rows.Scan(&it.ID, &it.FlagsetID, &it.Name, &it.CurrentVersion, &it.Disabled, &it.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, it)
	}
	return out, total, rows.Err()
}

func (r *FlagRepo) UpdateCurrentVersion(ctx context.Context, tx pgx.Tx, flagID, versionID string) error {
	_, err := tx.Exec(ctx,
		`UPDATE flags SET current_version_id = $2, updated_at = NOW() WHERE id = $1`, flagID, versionID)
	return err
}

func (r *FlagRepo) SetDisabled(ctx context.Context, flagID string, disabled bool) error {
	ct, err := r.db.Pool.Exec(ctx,
		`UPDATE flags SET disabled = $2, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, flagID, disabled)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *FlagRepo) SoftDelete(ctx context.Context, flagID string) error {
	ct, err := r.db.Pool.Exec(ctx,
		`UPDATE flags SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, flagID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *FlagRepo) GetFlagsetID(ctx context.Context, flagID string) (string, error) {
	var fsID string
	err := r.db.Pool.QueryRow(ctx,
		`SELECT flagset_id FROM flags WHERE id = $1 AND deleted_at IS NULL`, flagID).Scan(&fsID)
	if err != nil {
		if isNoRows(err) {
			return "", ErrNotFound
		}
		return "", err
	}
	return fsID, nil
}
