package catalog

import (
	"context"

	domaincatalog "github.com/yersonct/iam-service/internal/domain/catalog"
)

type UpdateFeatureInput struct {
	ID          string
	Name        string
	Description *string
	ActionLevel domaincatalog.ActionLevel
	IsActive    bool
}

type UpdateFeatureUseCase struct {
	featureRepo domaincatalog.FeatureRepository
}

func NewUpdateFeatureUseCase(featureRepo domaincatalog.FeatureRepository) *UpdateFeatureUseCase {
	return &UpdateFeatureUseCase{featureRepo: featureRepo}
}

func (uc *UpdateFeatureUseCase) Execute(ctx context.Context, in UpdateFeatureInput) error {
	if !in.ActionLevel.Valid() {
		return domaincatalog.ErrInvalidActionLevel
	}

	return uc.featureRepo.Update(ctx, in.ID, in.Name, in.Description, in.ActionLevel, in.IsActive)
}