package scopeoverride

import (
	"context"

	domainscopeoverride "github.com/yersonct/iam-service/internal/domain/scopeoverride"
)

type RemoveOverrideUseCase struct {
	repo domainscopeoverride.Repository
}

func NewRemoveOverrideUseCase(repo domainscopeoverride.Repository) *RemoveOverrideUseCase {
	return &RemoveOverrideUseCase{repo: repo}
}

func (uc *RemoveOverrideUseCase) Execute(ctx context.Context, id string) error {
	return uc.repo.Remove(ctx, id)
}