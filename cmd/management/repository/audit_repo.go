package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/thomaspoignant/go-feature-flag/cmd/management/model"
)

type AuditRepo struct{ db *DB }

func NewAuditRepo(db *DB) *AuditRepo { return &AuditRepo{db: db} }

type AuditInput struct {
	ActorUserID *string
	TeamID      *string
	FlagsetID   *string
	FlagID      *string
	Action      string
	TargetType  string
	TargetID    string
	Before      any
	After       any
	Metadata    any
}

func (r *AuditRepo) Record(ctx context.Context, tx pgx.Tx, in AuditInput) error {
	before := marshalAny(in.Before)
	after := marshalAny(in.After)
	meta := marshalAny(in.Metadata)

	q := `INSERT INTO audit_log (actor_user_id, team_id, flagset_id, flag_id, action, target_type, target_id, before, after, metadata)
	      VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	if tx != nil {
		_, err := tx.Exec(ctx, q, in.ActorUserID, in.TeamID, in.FlagsetID, in.FlagID, in.Action, in.TargetType, in.TargetID, before, after, meta)
		return err
	}
	_, err := r.db.Pool.Exec(ctx, q, in.ActorUserID, in.TeamID, in.FlagsetID, in.FlagID, in.Action, in.TargetType, in.TargetID, before, after, meta)
	return err
}

func marshalAny(v any) any {
	if v == nil {
		return nil
	}
	if raw, ok := v.(json.RawMessage); ok {
		if len(raw) == 0 {
			return nil
		}
		return []byte(raw)
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return b
}

func (r *AuditRepo) List(ctx context.Context, f model.AuditFilters) ([]model.AuditEntry, int, error) {
	qb := NewQueryBuilder()
	if f.TeamID != nil {
		qb.Add("a.team_id = $%d", *f.TeamID)
	}
	if f.FlagsetID != nil {
		qb.Add("a.flagset_id = $%d", *f.FlagsetID)
	}
	if f.FlagID != nil {
		qb.Add("a.flag_id = $%d", *f.FlagID)
	}
	if f.ActorID != nil {
		qb.Add("a.actor_user_id = $%d", *f.ActorID)
	}
	if f.Action != nil {
		qb.Add("a.action = $%d", *f.Action)
	}
	if f.From != nil {
		qb.Add("a.occurred_at >= $%d", *f.From)
	}
	if f.To != nil {
		qb.Add("a.occurred_at <= $%d", *f.To)
	}
	where, args := qb.Build()
	whereClause := "WHERE 1=1" + where

	var total int
	if err := r.db.Pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM audit_log a "+whereClause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 200 {
		f.PageSize = 50
	}
	offset := (f.Page - 1) * f.PageSize
	listQ := fmt.Sprintf(`
		SELECT a.id, a.occurred_at, a.actor_user_id, u.email AS actor_email,
		       a.team_id, a.flagset_id, a.flag_id, a.action, a.target_type, a.target_id,
		       a.before, a.after, a.metadata
		FROM audit_log a
		LEFT JOIN users u ON u.id = a.actor_user_id
		%s
		ORDER BY a.occurred_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, len(args)+1, len(args)+2)
	args = append(args, f.PageSize, offset)

	rows, err := r.db.Pool.Query(ctx, listQ, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []model.AuditEntry
	for rows.Next() {
		var e model.AuditEntry
		if err := rows.Scan(&e.ID, &e.OccurredAt, &e.ActorUserID, &e.ActorEmail,
			&e.TeamID, &e.FlagsetID, &e.FlagID, &e.Action, &e.TargetType, &e.TargetID,
			&e.Before, &e.After, &e.Metadata); err != nil {
			return nil, 0, err
		}
		out = append(out, e)
	}
	return out, total, rows.Err()
}
