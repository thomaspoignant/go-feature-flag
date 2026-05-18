package repository

import (
	"context"

	"github.com/thomaspoignant/go-feature-flag/cmd/management/model"
)

type TeamRepo struct{ db *DB }

func NewTeamRepo(db *DB) *TeamRepo { return &TeamRepo{db: db} }

func (r *TeamRepo) Create(ctx context.Context, name, description string) (*model.Team, error) {
	var t model.Team
	err := r.db.Pool.QueryRow(ctx,
		`INSERT INTO teams (name, description) VALUES ($1, $2)
		 RETURNING id, name, description, created_at, updated_at`,
		name, description,
	).Scan(&t.ID, &t.Name, &t.Description, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TeamRepo) GetByID(ctx context.Context, id string) (*model.Team, error) {
	var t model.Team
	err := r.db.Pool.QueryRow(ctx,
		`SELECT id, name, description, created_at, updated_at FROM teams WHERE id = $1`, id,
	).Scan(&t.ID, &t.Name, &t.Description, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (r *TeamRepo) List(ctx context.Context, userID string, superAdmin bool) ([]model.Team, error) {
	if superAdmin {
		return r.queryTeams(ctx, `SELECT id, name, description, created_at, updated_at FROM teams ORDER BY name`)
	}
	return r.queryTeams(ctx,
		`SELECT t.id, t.name, t.description, t.created_at, t.updated_at
		 FROM teams t
		 JOIN team_members tm ON tm.team_id = t.id
		 WHERE tm.user_id = $1
		 ORDER BY t.name`, userID)
}

func (r *TeamRepo) queryTeams(ctx context.Context, q string, args ...any) ([]model.Team, error) {
	rows, err := r.db.Pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Team
	for rows.Next() {
		var t model.Team
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *TeamRepo) Update(ctx context.Context, id string, name, description *string) (*model.Team, error) {
	var t model.Team
	err := r.db.Pool.QueryRow(ctx, `
		UPDATE teams SET
			name = COALESCE($2, name),
			description = COALESCE($3, description),
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, description, created_at, updated_at`, id, name, description,
	).Scan(&t.ID, &t.Name, &t.Description, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *TeamRepo) Delete(ctx context.Context, id string) error {
	ct, err := r.db.Pool.Exec(ctx, `DELETE FROM teams WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *TeamRepo) GetRole(ctx context.Context, teamID, userID string) (model.Role, error) {
	var role model.Role
	err := r.db.Pool.QueryRow(ctx,
		`SELECT role FROM team_members WHERE team_id = $1 AND user_id = $2`, teamID, userID,
	).Scan(&role)
	if err != nil {
		if isNoRows(err) {
			return "", nil
		}
		return "", err
	}
	return role, nil
}

func (r *TeamRepo) ListMembers(ctx context.Context, teamID string) ([]model.TeamMember, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT tm.team_id, tm.user_id, u.email, u.name, tm.role, tm.created_at
		FROM team_members tm
		JOIN users u ON u.id = tm.user_id
		WHERE tm.team_id = $1
		ORDER BY u.email`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.TeamMember
	for rows.Next() {
		var m model.TeamMember
		if err := rows.Scan(&m.TeamID, &m.UserID, &m.Email, &m.Name, &m.Role, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *TeamRepo) AddMember(ctx context.Context, teamID, userID string, role model.Role) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, $3)
		ON CONFLICT (team_id, user_id) DO UPDATE SET role = EXCLUDED.role`,
		teamID, userID, role)
	return err
}

func (r *TeamRepo) UpdateMemberRole(ctx context.Context, teamID, userID string, role model.Role) error {
	ct, err := r.db.Pool.Exec(ctx,
		`UPDATE team_members SET role = $3 WHERE team_id = $1 AND user_id = $2`, teamID, userID, role)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *TeamRepo) RemoveMember(ctx context.Context, teamID, userID string) error {
	ct, err := r.db.Pool.Exec(ctx,
		`DELETE FROM team_members WHERE team_id = $1 AND user_id = $2`, teamID, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *TeamRepo) Memberships(ctx context.Context, userID string) ([]model.TeamMembership, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT t.id, t.name, tm.role
		FROM team_members tm JOIN teams t ON t.id = tm.team_id
		WHERE tm.user_id = $1
		ORDER BY t.name`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.TeamMembership
	for rows.Next() {
		var m model.TeamMembership
		if err := rows.Scan(&m.TeamID, &m.TeamName, &m.Role); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
