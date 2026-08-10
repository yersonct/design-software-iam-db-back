package scopeoverride

import (
	"context"

	domainscopeoverride "github.com/yersonct/iam-service/internal/domain/scopeoverride"
)

type ListOverridesUseCase struct {
	repo domainscopeoverride.Repository
}

func NewListOverridesUseCase(repo domainscopeoverride.Repository) *ListOverridesUseCase {
	return &ListOverridesUseCase{repo: repo}
}

func (uc *ListOverridesUseCase) Execute(ctx context.Context, userID string) ([]domainscopeoverride.ScopeOverrideItem, error) {
	return uc.repo.ListByUser(ctx, userID)
}