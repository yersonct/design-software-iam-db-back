package catalog

import (
	"context"
	"fmt"

	domaincatalog "github.com/yersonct/iam-service/internal/domain/catalog"
)

type ListFeaturesByModuleUseCase struct {
	featureRepo domaincatalog.FeatureRepository
}

func NewListFeaturesByModuleUseCase(featureRepo domaincatalog.FeatureRepository) *ListFeaturesByModuleUseCase {
	return &ListFeaturesByModuleUseCase{featureRepo: featureRepo}
}

func (uc *ListFeaturesByModuleUseCase) Execute(ctx context.Context, moduleID string) ([]domaincatalog.Feature, error) {
	exists, err := uc.featureRepo.ModuleExists(ctx, moduleID)
	if err != nil {
		return nil, fmt.Errorf("check module existence: %w", err)
	}
	if !exists {
		return nil, domaincatalog.ErrModuleNotFound
	}

	return uc.featureRepo.ListByModuleID(ctx, moduleID)
}