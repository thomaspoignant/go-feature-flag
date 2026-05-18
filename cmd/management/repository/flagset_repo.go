package repository

import (
	"context"
	"encoding/json"

	"github.com/thomaspoignant/go-feature-flag/cmd/management/model"
)

type FlagsetRepo struct{ db *DB }

func NewFlagsetRepo(db *DB) *FlagsetRepo { return &FlagsetRepo{db: db} }

func (r *FlagsetRepo) Create(ctx context.Context, teamID string, req *model.CreateFlagsetRequest) (*model.Flagset, error) {
	var fs model.Flagset
	err := r.db.Pool.QueryRow(ctx, `
		INSERT INTO flagsets (team_id, name, description, polling_interval_ms, file_format, retrievers, exporters, notifiers)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, team_id, name, description, api_key_hashes, polling_interval_ms, file_format, retrievers, exporters, notifiers, created_at, updated_at`,
		teamID, req.Name, req.Description, req.PollingIntervalMs, req.FileFormat,
		nullableJSON(req.Retrievers), nullableJSON(req.Exporters), nullableJSON(req.Notifiers),
	).Scan(
		&fs.ID, &fs.TeamID, &fs.Name, &fs.Description, &fs.APIKeyHashes, &fs.PollingIntervalMs, &fs.FileFormat,
		&fs.Retrievers, &fs.Exporters, &fs.Notifiers, &fs.CreatedAt, &fs.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	fs.APIKeyCount = len(fs.APIKeyHashes)
	return &fs, nil
}

func (r *FlagsetRepo) GetByID(ctx context.Context, id string) (*model.Flagset, error) {
	var fs model.Flagset
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, team_id, name, description, api_key_hashes, polling_interval_ms, file_format, retrievers, exporters, notifiers, created_at, updated_at
		FROM flagsets WHERE id = $1`, id,
	).Scan(
		&fs.ID, &fs.TeamID, &fs.Name, &fs.Description, &fs.APIKeyHashes, &fs.PollingIntervalMs, &fs.FileFormat,
		&fs.Retrievers, &fs.Exporters, &fs.Notifiers, &fs.CreatedAt, &fs.UpdatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return nil, nil
		}
		return nil, err
	}
	fs.APIKeyCount = len(fs.APIKeyHashes)
	return &fs, nil
}

func (r *FlagsetRepo) ListByTeam(ctx context.Context, teamID string) ([]model.Flagset, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, team_id, name, description, api_key_hashes, polling_interval_ms, file_format, retrievers, exporters, notifiers, created_at, updated_at
		FROM flagsets WHERE team_id = $1 ORDER BY name`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Flagset
	for rows.Next() {
		var fs model.Flagset
		if err := rows.Scan(
			&fs.ID, &fs.TeamID, &fs.Name, &fs.Description, &fs.APIKeyHashes, &fs.PollingIntervalMs, &fs.FileFormat,
			&fs.Retrievers, &fs.Exporters, &fs.Notifiers, &fs.CreatedAt, &fs.UpdatedAt,
		); err != nil {
			return nil, err
		}
		fs.APIKeyCount = len(fs.APIKeyHashes)
		out = append(out, fs)
	}
	return out, rows.Err()
}

func (r *FlagsetRepo) Update(ctx context.Context, id string, req *model.UpdateFlagsetRequest) (*model.Flagset, error) {
	var fs model.Flagset
	err := r.db.Pool.QueryRow(ctx, `
		UPDATE flagsets SET
			name = COALESCE($2, name),
			description = COALESCE($3, description),
			polling_interval_ms = COALESCE($4, polling_interval_ms),
			file_format = COALESCE($5, file_format),
			retrievers = COALESCE($6, retrievers),
			exporters = COALESCE($7, exporters),
			notifiers = COALESCE($8, notifiers),
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, team_id, name, description, api_key_hashes, polling_interval_ms, file_format, retrievers, exporters, notifiers, created_at, updated_at`,
		id, req.Name, req.Description, req.PollingIntervalMs, req.FileFormat,
		nullableJSON(req.Retrievers), nullableJSON(req.Exporters), nullableJSON(req.Notifiers),
	).Scan(
		&fs.ID, &fs.TeamID, &fs.Name, &fs.Description, &fs.APIKeyHashes, &fs.PollingIntervalMs, &fs.FileFormat,
		&fs.Retrievers, &fs.Exporters, &fs.Notifiers, &fs.CreatedAt, &fs.UpdatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	fs.APIKeyCount = len(fs.APIKeyHashes)
	return &fs, nil
}

func (r *FlagsetRepo) Delete(ctx context.Context, id string) error {
	ct, err := r.db.Pool.Exec(ctx, `DELETE FROM flagsets WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *FlagsetRepo) AddAPIKeyHash(ctx context.Context, id, hash string) error {
	ct, err := r.db.Pool.Exec(ctx,
		`UPDATE flagsets SET api_key_hashes = array_append(api_key_hashes, $2), updated_at = NOW() WHERE id = $1`,
		id, hash)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *FlagsetRepo) RemoveAPIKeyHash(ctx context.Context, id, hash string) error {
	ct, err := r.db.Pool.Exec(ctx,
		`UPDATE flagsets SET api_key_hashes = array_remove(api_key_hashes, $2), updated_at = NOW() WHERE id = $1`,
		id, hash)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func nullableJSON(raw json.RawMessage) any {
	if len(raw) == 0 {
		return nil
	}
	return []byte(raw)
}
