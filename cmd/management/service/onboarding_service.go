package service

import (
	"context"

	"github.com/thomaspoignant/go-feature-flag/cmd/management/model"
	"github.com/thomaspoignant/go-feature-flag/cmd/management/repository"
)

type OnboardingService struct {
	teams *repository.TeamRepo
	audit *repository.AuditRepo
}

func NewOnboardingService(teams *repository.TeamRepo, audit *repository.AuditRepo) *OnboardingService {
	return &OnboardingService{teams: teams, audit: audit}
}

// CreateTeamForNewUser lets a user with no team memberships create a team and
// be added as its admin. Rejects users already attached to any team.
func (s *OnboardingService) CreateTeamForNewUser(ctx context.Context, userID string, req *model.CreateTeamRequest) (*model.Team, error) {
	if req.Name == "" {
		return nil, model.ValidationErrors{{Field: "name", Message: "name is required"}}
	}
	memberships, err := s.teams.Memberships(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(memberships) > 0 {
		return nil, ErrForbidden
	}
	t, err := s.teams.Create(ctx, req.Name, req.Description)
	if err != nil {
		return nil, err
	}
	if err := s.teams.AddMember(ctx, t.ID, userID, model.RoleAdmin); err != nil {
		return nil, err
	}
	_ = s.audit.Record(ctx, nil, repository.AuditInput{
		ActorUserID: &userID, TeamID: &t.ID,
		Action: "team.onboarding.create", TargetType: "team", TargetID: t.ID, After: t,
	})
	return t, nil
}
