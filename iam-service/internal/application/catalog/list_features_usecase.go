package catalog

import (
	"context"

	domaincatalog "github.com/yersonct/iam-service/internal/domain/catalog"
)

type ListFeaturesUseCase struct {
	featureRepo domaincatalog.FeatureRepository
}

func NewListFeaturesUseCase(featureRepo domaincatalog.FeatureRepository) *ListFeaturesUseCase {
	return &ListFeaturesUseCase{featureRepo: featureRepo}
}

func (uc *ListFeaturesUseCase) Execute(ctx context.Context) ([]domaincatalog.Feature, error) {
	return uc.featureRepo.List(ctx)
}

// --- Segundo caso de uso en el mismo archivo: agrupadas por módulo ---
// (Lo separo abajo por claridad, pero puedes ponerlo en su propio archivo
// list_features_by_module_usecase.go si prefieres 1 archivo = 1 caso de uso)