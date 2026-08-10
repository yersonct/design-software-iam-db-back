package rolefeature

import (
	"context"

	domainrolefeature "github.com/yersonct/iam-service/internal/domain/rolefeature"
)

type ListRoleFeaturesUseCase struct {
	repo domainrolefeature.Repository
}

func NewListRoleFeaturesUseCase(repo domainrolefeature.Repository) *ListRoleFeaturesUseCase {
	return &ListRoleFeaturesUseCase{repo: repo}
}

func (uc *ListRoleFeaturesUseCase) Execute(ctx context.Context, roleID string) ([]domainrolefeature.RoleFeatureItem, error) {
	return uc.repo.ListByRole(ctx, roleID)
}