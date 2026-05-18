package repository

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"

	"github.com/thomaspoignant/go-feature-flag/cmd/management/model"
)

type FlagVersionRepo struct{ db *DB }

func NewFlagVersionRepo(db *DB) *FlagVersionRepo { return &FlagVersionRepo{db: db} }

func (r *FlagVersionRepo) Insert(ctx context.Context, tx pgx.Tx, flagID string, payload json.RawMessage, comment string, createdBy *string) (*model.FlagVersion, error) {
	var nextNum int
	err := tx.QueryRow(ctx,
		`SELECT COALESCE(MAX(version_number),0)+1 FROM flag_versions WHERE flag_id = $1`, flagID).Scan(&nextNum)
	if err != nil {
		return nil, err
	}
	var v model.FlagVersion
	err = tx.QueryRow(ctx, `
		INSERT INTO flag_versions (flag_id, version_number, payload, comment, created_by)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id, flag_id, version_number, payload, comment, created_by, created_at`,
		flagID, nextNum, []byte(payload), comment, createdBy,
	).Scan(&v.ID, &v.FlagID, &v.VersionNumber, &v.Payload, &v.Comment, &v.CreatedBy, &v.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *FlagVersionRepo) GetByFlagAndNumber(ctx context.Context, flagID string, num int) (*model.FlagVersion, error) {
	var v model.FlagVersion
	err := r.db.Pool.QueryRow(ctx, `
		SELECT fv.id, fv.flag_id, fv.version_number, fv.payload, fv.comment, fv.created_by, fv.created_at,
		       u.name AS created_by_name
		FROM flag_versions fv
		LEFT JOIN users u ON u.id = fv.created_by
		WHERE fv.flag_id = $1 AND fv.version_number = $2`,
		flagID, num,
	).Scan(&v.ID, &v.FlagID, &v.VersionNumber, &v.Payload, &v.Comment, &v.CreatedBy, &v.CreatedAt, &v.CreatedByName)
	if err != nil {
		if isNoRows(err) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *FlagVersionRepo) ListByFlag(ctx context.Context, flagID string, limit, offset int) ([]model.FlagVersion, int, error) {
	var total int
	if err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM flag_versions WHERE flag_id = $1`, flagID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Pool.Query(ctx, `
		SELECT fv.id, fv.flag_id, fv.version_number, fv.payload, fv.comment, fv.created_by, fv.created_at,
		       u.name AS created_by_name
		FROM flag_versions fv
		LEFT JOIN users u ON u.id = fv.created_by
		WHERE fv.flag_id = $1
		ORDER BY fv.version_number DESC
		LIMIT $2 OFFSET $3`, flagID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []model.FlagVersion
	for rows.Next() {
		var v model.FlagVersion
		if err := rows.Scan(&v.ID, &v.FlagID, &v.VersionNumber, &v.Payload, &v.Comment, &v.CreatedBy, &v.CreatedAt, &v.CreatedByName); err != nil {
			return nil, 0, err
		}
		out = append(out, v)
	}
	return out, total, rows.Err()
}
