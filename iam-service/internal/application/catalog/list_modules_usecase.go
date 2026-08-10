package catalog

import (
	"context"

	domaincatalog "github.com/yersonct/iam-service/internal/domain/catalog"
)

type ListModulesUseCase struct {
	moduleRepo domaincatalog.ModuleRepository
}

func NewListModulesUseCase(moduleRepo domaincatalog.ModuleRepository) *ListModulesUseCase {
	return &ListModulesUseCase{moduleRepo: moduleRepo}
}

func (uc *ListModulesUseCase) Execute(ctx context.Context) ([]domaincatalog.Module, error) {
	return uc.moduleRepo.List(ctx)
}