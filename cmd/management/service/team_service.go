package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/thomaspoignant/go-feature-flag/cmd/management/model"
	"github.com/thomaspoignant/go-feature-flag/cmd/management/repository"
)

var (
	ErrForbidden = errors.New("forbidden")
	ErrNotFound  = errors.New("not found")
	ErrConflict  = errors.New("conflict")
)

type TeamService struct {
	teams *repository.TeamRepo
	users *repository.UserRepo
	audit *repository.AuditRepo
}

func NewTeamService(teams *repository.TeamRepo, users *repository.UserRepo, audit *repository.AuditRepo) *TeamService {
	return &TeamService{teams: teams, users: users, audit: audit}
}

func (s *TeamService) Create(ctx context.Context, actorID string, req *model.CreateTeamRequest) (*model.Team, error) {
	if req.Name == "" {
		return nil, model.ValidationErrors{{Field: "name", Message: "name is required"}}
	}
	t, err := s.teams.Create(ctx, req.Name, req.Description)
	if err != nil {
		return nil, err
	}
	_ = s.audit.Record(ctx, nil, repository.AuditInput{
		ActorUserID: &actorID, TeamID: &t.ID,
		Action: "team.create", TargetType: "team", TargetID: t.ID, After: t,
	})
	return t, nil
}

func (s *TeamService) List(ctx context.Context, userID string, superAdmin bool) ([]model.Team, error) {
	return s.teams.List(ctx, userID, superAdmin)
}

func (s *TeamService) Get(ctx context.Context, id string) (*model.Team, error) {
	return s.teams.GetByID(ctx, id)
}

func (s *TeamService) Update(ctx context.Context, actorID, id string, req *model.UpdateTeamRequest) (*model.Team, error) {
	before, err := s.teams.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if before == nil {
		return nil, ErrNotFound
	}
	t, err := s.teams.Update(ctx, id, req.Name, req.Description)
	if err != nil {
		return nil, err
	}
	_ = s.audit.Record(ctx, nil, repository.AuditInput{
		ActorUserID: &actorID, TeamID: &t.ID,
		Action: "team.update", TargetType: "team", TargetID: t.ID, Before: before, After: t,
	})
	return t, nil
}

func (s *TeamService) Delete(ctx context.Context, actorID, id string) error {
	before, err := s.teams.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if before == nil {
		return ErrNotFound
	}
	if err := s.teams.Delete(ctx, id); err != nil {
		return err
	}
	_ = s.audit.Record(ctx, nil, repository.AuditInput{
		ActorUserID: &actorID, TeamID: &id,
		Action: "team.delete", TargetType: "team", TargetID: id, Before: before,
	})
	return nil
}

func (s *TeamService) ListMembers(ctx context.Context, teamID string) ([]model.TeamMember, error) {
	return s.teams.ListMembers(ctx, teamID)
}

func (s *TeamService) AddMember(ctx context.Context, actorID, teamID string, req *model.AddMemberRequest) (*model.TeamMember, error) {
	if !req.Role.IsValid() {
		return nil, model.ValidationErrors{{Field: "role", Message: "invalid role"}}
	}
	u, err := s.users.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, fmt.Errorf("%w: user with email %s not found", ErrNotFound, req.Email)
	}
	if err := s.teams.AddMember(ctx, teamID, u.ID, req.Role); err != nil {
		return nil, err
	}
	m := &model.TeamMember{TeamID: teamID, UserID: u.ID, Email: u.Email, Name: u.Name, Role: req.Role}
	_ = s.audit.Record(ctx, nil, repository.AuditInput{
		ActorUserID: &actorID, TeamID: &teamID,
		Action: "team.member.add", TargetType: "user", TargetID: u.ID, After: m,
	})
	return m, nil
}

func (s *TeamService) UpdateMember(ctx context.Context, actorID, teamID, userID string, role model.Role) error {
	if !role.IsValid() {
		return model.ValidationErrors{{Field: "role", Message: "invalid role"}}
	}
	if err := s.teams.UpdateMemberRole(ctx, teamID, userID, role); err != nil {
		return err
	}
	_ = s.audit.Record(ctx, nil, repository.AuditInput{
		ActorUserID: &actorID, TeamID: &teamID,
		Action: "team.member.update", TargetType: "user", TargetID: userID,
		After: map[string]any{"role": role},
	})
	return nil
}

func (s *TeamService) RemoveMember(ctx context.Context, actorID, teamID, userID string) error {
	if err := s.teams.RemoveMember(ctx, teamID, userID); err != nil {
		return err
	}
	_ = s.audit.Record(ctx, nil, repository.AuditInput{
		ActorUserID: &actorID, TeamID: &teamID,
		Action: "team.member.remove", TargetType: "user", TargetID: userID,
	})
	return nil
}

func (s *TeamService) Role(ctx context.Context, teamID, userID string) (model.Role, error) {
	return s.teams.GetRole(ctx, teamID, userID)
}

func (s *TeamService) Memberships(ctx context.Context, userID string) ([]model.TeamMembership, error) {
	return s.teams.Memberships(ctx, userID)
}
