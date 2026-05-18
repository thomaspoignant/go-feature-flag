package service

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"

	"github.com/thomaspoignant/go-feature-flag/cmd/management/model"
	"github.com/thomaspoignant/go-feature-flag/cmd/management/repository"
)

type FlagService struct {
	db       *repository.DB
	flags    *repository.FlagRepo
	versions *repository.FlagVersionRepo
	flagsets *repository.FlagsetRepo
	audit    *repository.AuditRepo
}

func NewFlagService(db *repository.DB, flags *repository.FlagRepo, versions *repository.FlagVersionRepo, flagsets *repository.FlagsetRepo, audit *repository.AuditRepo) *FlagService {
	return &FlagService{db: db, flags: flags, versions: versions, flagsets: flagsets, audit: audit}
}

func (s *FlagService) Create(ctx context.Context, actorID, flagsetID string, req *model.CreateFlagRequest) (*model.Flag, error) {
	if errs := model.ValidateFlagName(req.Name); len(errs) > 0 {
		return nil, errs
	}
	if errs := model.ValidateFlagPayload(req.Payload); len(errs) > 0 {
		return nil, errs
	}
	_ = req.Payload.Convert()
	fs, err := s.flagsets.GetByID(ctx, flagsetID)
	if err != nil {
		return nil, err
	}
	if fs == nil {
		return nil, ErrNotFound
	}
	payload, err := json.Marshal(req.Payload)
	if err != nil {
		return nil, err
	}

	var out *model.Flag
	err = s.db.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		f, err := s.flags.Create(ctx, tx, flagsetID, req.Name)
		if err != nil {
			return err
		}
		v, err := s.versions.Insert(ctx, tx, f.ID, payload, req.Comment, &actorID)
		if err != nil {
			return err
		}
		if err := s.flags.UpdateCurrentVersion(ctx, tx, f.ID, v.ID); err != nil {
			return err
		}
		f.CurrentVersionID = &v.ID
		f.CurrentVersion = &v.VersionNumber
		f.Payload = v.Payload
		out = f
		return s.audit.Record(ctx, tx, repository.AuditInput{
			ActorUserID: &actorID, TeamID: &fs.TeamID, FlagsetID: &flagsetID, FlagID: &f.ID,
			Action: "flag.create", TargetType: "flag", TargetID: f.ID, After: f,
		})
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *FlagService) Get(ctx context.Context, id string) (*model.Flag, error) {
	return s.flags.GetByID(ctx, id)
}

func (s *FlagService) List(ctx context.Context, flagsetID string, f model.FlagFilters) ([]model.FlagListItem, int, error) {
	return s.flags.ListByFlagset(ctx, flagsetID, f)
}

func (s *FlagService) Update(ctx context.Context, actorID, id string, req *model.UpdateFlagRequest) (*model.Flag, error) {
	if errs := model.ValidateFlagPayload(req.Payload); len(errs) > 0 {
		return nil, errs
	}
	before, err := s.flags.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if before == nil {
		return nil, ErrNotFound
	}
	fs, err := s.flagsets.GetByID(ctx, before.FlagsetID)
	if err != nil {
		return nil, err
	}
	payload, err := json.Marshal(req.Payload)
	if err != nil {
		return nil, err
	}

	var out *model.Flag
	err = s.db.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		v, err := s.versions.Insert(ctx, tx, id, payload, req.Comment, &actorID)
		if err != nil {
			return err
		}
		if err := s.flags.UpdateCurrentVersion(ctx, tx, id, v.ID); err != nil {
			return err
		}
		after := *before
		after.CurrentVersionID = &v.ID
		after.CurrentVersion = &v.VersionNumber
		after.Payload = v.Payload
		out = &after
		return s.audit.Record(ctx, tx, repository.AuditInput{
			ActorUserID: &actorID, TeamID: &fs.TeamID, FlagsetID: &before.FlagsetID, FlagID: &id,
			Action: "flag.update", TargetType: "flag", TargetID: id, Before: before, After: &after,
		})
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *FlagService) SetDisabled(ctx context.Context, actorID, id string, disabled bool) error {
	before, err := s.flags.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if before == nil {
		return ErrNotFound
	}
	if err := s.flags.SetDisabled(ctx, id, disabled); err != nil {
		return err
	}
	fs, _ := s.flagsets.GetByID(ctx, before.FlagsetID)
	var teamID *string
	if fs != nil {
		teamID = &fs.TeamID
	}
	_ = s.audit.Record(ctx, nil, repository.AuditInput{
		ActorUserID: &actorID, TeamID: teamID, FlagsetID: &before.FlagsetID, FlagID: &id,
		Action: "flag.disable", TargetType: "flag", TargetID: id,
		Before: map[string]any{"disabled": before.Disabled},
		After:  map[string]any{"disabled": disabled},
	})
	return nil
}

func (s *FlagService) Delete(ctx context.Context, actorID, id string) error {
	before, err := s.flags.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if before == nil {
		return ErrNotFound
	}
	if err := s.flags.SoftDelete(ctx, id); err != nil {
		return err
	}
	fs, _ := s.flagsets.GetByID(ctx, before.FlagsetID)
	var teamID *string
	if fs != nil {
		teamID = &fs.TeamID
	}
	_ = s.audit.Record(ctx, nil, repository.AuditInput{
		ActorUserID: &actorID, TeamID: teamID, FlagsetID: &before.FlagsetID, FlagID: &id,
		Action: "flag.delete", TargetType: "flag", TargetID: id, Before: before,
	})
	return nil
}

func (s *FlagService) FlagsetID(ctx context.Context, flagID string) (string, error) {
	return s.flags.GetFlagsetID(ctx, flagID)
}
