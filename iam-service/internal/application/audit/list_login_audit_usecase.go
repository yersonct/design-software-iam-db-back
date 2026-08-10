package audit

import (
	"context"

	domainaudit "github.com/yersonct/iam-service/internal/domain/audit"
)

type ListLoginAuditUseCase struct {
	repo domainaudit.Repository
}

func NewListLoginAuditUseCase(repo domainaudit.Repository) *ListLoginAuditUseCase {
	return &ListLoginAuditUseCase{repo: repo}
}

func (uc *ListLoginAuditUseCase) Execute(ctx context.Context, filter domainaudit.ListFilter) (*domainaudit.ListResult, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}
	return uc.repo.ListLoginAttempts(ctx, filter)
}