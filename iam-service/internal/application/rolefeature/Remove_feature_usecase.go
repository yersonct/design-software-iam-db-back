package rolefeature

import (
	"context"

	domainrolefeature "github.com/yersonct/iam-service/internal/domain/rolefeature"
)

type RemoveFeatureUseCase struct {
	repo domainrolefeature.Repository
}

func NewRemoveFeatureUseCase(repo domainrolefeature.Repository) *RemoveFeatureUseCase {
	return &RemoveFeatureUseCase{repo: repo}
}

func (uc *RemoveFeatureUseCase) Execute(ctx context.Context, roleID string, featureID string) error {
	return uc.repo.Remove(ctx, roleID, featureID)
}