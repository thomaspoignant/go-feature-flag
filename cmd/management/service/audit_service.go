package service

import (
	"context"

	"github.com/thomaspoignant/go-feature-flag/cmd/management/model"
	"github.com/thomaspoignant/go-feature-flag/cmd/management/repository"
)

type AuditService struct {
	repo *repository.AuditRepo
}

func NewAuditService(repo *repository.AuditRepo) *AuditService { return &AuditService{repo: repo} }

func (s *AuditService) List(ctx context.Context, f model.AuditFilters) ([]model.AuditEntry, int, error) {
	return s.repo.List(ctx, f)
}
