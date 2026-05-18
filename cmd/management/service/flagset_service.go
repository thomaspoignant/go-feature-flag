package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"

	"github.com/thomaspoignant/go-feature-flag/cmd/management/model"
	"github.com/thomaspoignant/go-feature-flag/cmd/management/repository"
)

type FlagsetService struct {
	repo  *repository.FlagsetRepo
	audit *repository.AuditRepo
}

func NewFlagsetService(repo *repository.FlagsetRepo, audit *repository.AuditRepo) *FlagsetService {
	return &FlagsetService{repo: repo, audit: audit}
}

func (s *FlagsetService) Create(ctx context.Context, actorID, teamID string, req *model.CreateFlagsetRequest) (*model.Flagset, error) {
	if req.Name == "" {
		return nil, model.ValidationErrors{{Field: "name", Message: "name is required"}}
	}
	fs, err := s.repo.Create(ctx, teamID, req)
	if err != nil {
		return nil, err
	}
	_ = s.audit.Record(ctx, nil, repository.AuditInput{
		ActorUserID: &actorID, TeamID: &teamID, FlagsetID: &fs.ID,
		Action: "flagset.create", TargetType: "flagset", TargetID: fs.ID, After: fs,
	})
	return fs, nil
}

func (s *FlagsetService) Get(ctx context.Context, id string) (*model.Flagset, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *FlagsetService) ListByTeam(ctx context.Context, teamID string) ([]model.Flagset, error) {
	return s.repo.ListByTeam(ctx, teamID)
}

func (s *FlagsetService) Update(ctx context.Context, actorID, id string, req *model.UpdateFlagsetRequest) (*model.Flagset, error) {
	before, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if before == nil {
		return nil, ErrNotFound
	}
	fs, err := s.repo.Update(ctx, id, req)
	if err != nil {
		return nil, err
	}
	_ = s.audit.Record(ctx, nil, repository.AuditInput{
		ActorUserID: &actorID, TeamID: &fs.TeamID, FlagsetID: &fs.ID,
		Action: "flagset.update", TargetType: "flagset", TargetID: fs.ID, Before: before, After: fs,
	})
	return fs, nil
}

func (s *FlagsetService) Delete(ctx context.Context, actorID, id string) error {
	before, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if before == nil {
		return ErrNotFound
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	_ = s.audit.Record(ctx, nil, repository.AuditInput{
		ActorUserID: &actorID, TeamID: &before.TeamID, FlagsetID: &id,
		Action: "flagset.delete", TargetType: "flagset", TargetID: id, Before: before,
	})
	return nil
}

func (s *FlagsetService) CreateAPIKey(ctx context.Context, actorID, id string) (*model.CreateAPIKeyResponse, error) {
	fs, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if fs == nil {
		return nil, ErrNotFound
	}
	raw, err := randomKey(32)
	if err != nil {
		return nil, err
	}
	key := "goff_" + raw
	hash := sha256Hex(key)
	if err := s.repo.AddAPIKeyHash(ctx, id, hash); err != nil {
		return nil, err
	}
	_ = s.audit.Record(ctx, nil, repository.AuditInput{
		ActorUserID: &actorID, TeamID: &fs.TeamID, FlagsetID: &id,
		Action: "flagset.apikey.create", TargetType: "flagset", TargetID: id,
		Metadata: map[string]any{"keyHash": hash},
	})
	return &model.CreateAPIKeyResponse{APIKey: key, KeyHash: hash}, nil
}

func (s *FlagsetService) DeleteAPIKey(ctx context.Context, actorID, id, hash string) error {
	fs, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if fs == nil {
		return ErrNotFound
	}
	if err := s.repo.RemoveAPIKeyHash(ctx, id, hash); err != nil {
		return err
	}
	_ = s.audit.Record(ctx, nil, repository.AuditInput{
		ActorUserID: &actorID, TeamID: &fs.TeamID, FlagsetID: &id,
		Action: "flagset.apikey.delete", TargetType: "flagset", TargetID: id,
		Metadata: map[string]any{"keyHash": hash},
	})
	return nil
}

func randomKey(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
