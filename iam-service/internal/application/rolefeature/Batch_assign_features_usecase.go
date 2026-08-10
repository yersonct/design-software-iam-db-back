package rolefeature

import (
	"context"

	domainrolefeature "github.com/yersonct/iam-service/internal/domain/rolefeature"
)

type BatchAssignFeatureItem struct {
	FeatureID string
	ScopeType string
}

type BatchAssignFeaturesInput struct {
	RoleID   string
	Features []BatchAssignFeatureItem
}

type BatchAssignFeaturesUseCase struct {
	repo domainrolefeature.Repository
}

func NewBatchAssignFeaturesUseCase(repo domainrolefeature.Repository) *BatchAssignFeaturesUseCase {
	return &BatchAssignFeaturesUseCase{repo: repo}
}

func (uc *BatchAssignFeaturesUseCase) Execute(ctx context.Context, input BatchAssignFeaturesInput) error {
	items := make([]domainrolefeature.RoleFeature, 0, len(input.Features))

	for _, f := range input.Features {
		scope := domainrolefeature.ScopeType(f.ScopeType)
		if !scope.IsValid() {
			return domainrolefeature.ErrInvalidScopeType
		}
		items = append(items, domainrolefeature.RoleFeature{
			RoleID:    input.RoleID,
			FeatureID: f.FeatureID,
			ScopeType: scope,
		})
	}

	return uc.repo.ReplaceAll(ctx, input.RoleID, items)
}