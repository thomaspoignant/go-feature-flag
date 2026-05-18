package service

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/thomaspoignant/go-feature-flag/cmd/management/model"
	"github.com/thomaspoignant/go-feature-flag/cmd/management/repository"
)

type VersionService struct {
	db       *repository.DB
	flags    *repository.FlagRepo
	versions *repository.FlagVersionRepo
	flagsets *repository.FlagsetRepo
	audit    *repository.AuditRepo
}

func NewVersionService(db *repository.DB, flags *repository.FlagRepo, versions *repository.FlagVersionRepo, flagsets *repository.FlagsetRepo, audit *repository.AuditRepo) *VersionService {
	return &VersionService{db: db, flags: flags, versions: versions, flagsets: flagsets, audit: audit}
}

func (s *VersionService) List(ctx context.Context, flagID string, page, pageSize int) ([]model.FlagVersion, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 50
	}
	return s.versions.ListByFlag(ctx, flagID, pageSize, (page-1)*pageSize)
}

func (s *VersionService) Get(ctx context.Context, flagID string, num int) (*model.FlagVersion, error) {
	return s.versions.GetByFlagAndNumber(ctx, flagID, num)
}

func (s *VersionService) Rollback(ctx context.Context, actorID, flagID string, num int, comment string) (*model.FlagVersion, error) {
	src, err := s.versions.GetByFlagAndNumber(ctx, flagID, num)
	if err != nil {
		return nil, err
	}
	if src == nil {
		return nil, ErrNotFound
	}
	flag, err := s.flags.GetByID(ctx, flagID)
	if err != nil {
		return nil, err
	}
	if flag == nil {
		return nil, ErrNotFound
	}
	fs, _ := s.flagsets.GetByID(ctx, flag.FlagsetID)
	if comment == "" {
		comment = "rollback to version " + itoa(num)
	}

	var out *model.FlagVersion
	err = s.db.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		v, err := s.versions.Insert(ctx, tx, flagID, src.Payload, comment, &actorID)
		if err != nil {
			return err
		}
		if err := s.flags.UpdateCurrentVersion(ctx, tx, flagID, v.ID); err != nil {
			return err
		}
		out = v
		var teamID *string
		if fs != nil {
			teamID = &fs.TeamID
		}
		return s.audit.Record(ctx, tx, repository.AuditInput{
			ActorUserID: &actorID, TeamID: teamID, FlagsetID: &flag.FlagsetID, FlagID: &flagID,
			Action: "flag.rollback", TargetType: "flag", TargetID: flagID,
			Metadata: map[string]any{"fromVersion": num, "toVersion": v.VersionNumber},
		})
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
