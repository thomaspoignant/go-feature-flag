package repository

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/thomaspoignant/go-feature-flag/cmd/management/model"
)

type UserRepo struct{ db *DB }

func NewUserRepo(db *DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) GetByID(ctx context.Context, id string) (*model.User, error) {
	return r.scanOne(ctx, `SELECT id, email, name, oidc_sub, is_super_admin, created_at, updated_at, last_login_at FROM users WHERE id = $1`, id)
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	return r.scanOne(ctx, `SELECT id, email, name, oidc_sub, is_super_admin, created_at, updated_at, last_login_at FROM users WHERE email = $1`, email)
}

func (r *UserRepo) GetByOIDCSub(ctx context.Context, sub string) (*model.User, error) {
	return r.scanOne(ctx, `SELECT id, email, name, oidc_sub, is_super_admin, created_at, updated_at, last_login_at FROM users WHERE oidc_sub = $1`, sub)
}

func (r *UserRepo) scanOne(ctx context.Context, q string, args ...any) (*model.User, error) {
	var u model.User
	err := r.db.Pool.QueryRow(ctx, q, args...).Scan(
		&u.ID, &u.Email, &u.Name, &u.OIDCSub, &u.IsSuperAdmin,
		&u.CreatedAt, &u.UpdatedAt, &u.LastLoginAt,
	)
	if err != nil {
		if isNoRows(err) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) UpsertFromOIDC(ctx context.Context, tx pgx.Tx, email, name, sub string, isSuperAdmin bool) (*model.User, error) {
	q := `
		INSERT INTO users (email, name, oidc_sub, is_super_admin, last_login_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (email) DO UPDATE
		SET name = EXCLUDED.name,
		    oidc_sub = COALESCE(users.oidc_sub, EXCLUDED.oidc_sub),
		    is_super_admin = users.is_super_admin OR EXCLUDED.is_super_admin,
		    last_login_at = NOW(),
		    updated_at = NOW()
		RETURNING id, email, name, oidc_sub, is_super_admin, created_at, updated_at, last_login_at`
	var u model.User
	var execer interface {
		QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	} = r.db.Pool
	if tx != nil {
		execer = tx
	}
	err := execer.QueryRow(ctx, q, strings.ToLower(email), name, sub, isSuperAdmin).Scan(
		&u.ID, &u.Email, &u.Name, &u.OIDCSub, &u.IsSuperAdmin,
		&u.CreatedAt, &u.UpdatedAt, &u.LastLoginAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) MarkLogin(ctx context.Context, userID string) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE users SET last_login_at = $1, updated_at = NOW() WHERE id = $2`, time.Now(), userID)
	return err
}
