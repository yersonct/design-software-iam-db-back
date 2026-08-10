package catalog

import (
	"context"

	domaincatalog "github.com/yersonct/iam-service/internal/domain/catalog"
)

type UpdateModuleInput struct {
	ID           string
	Name         string
	Description  *string
	DisplayOrder int16
	IconKey      *string
	IsActive     bool
}

type UpdateModuleUseCase struct {
	moduleRepo domaincatalog.ModuleRepository
}

func NewUpdateModuleUseCase(moduleRepo domaincatalog.ModuleRepository) *UpdateModuleUseCase {
	return &UpdateModuleUseCase{moduleRepo: moduleRepo}
}

func (uc *UpdateModuleUseCase) Execute(ctx context.Context, in UpdateModuleInput) error {
	return uc.moduleRepo.Update(ctx, in.ID, in.Name, in.Description, in.DisplayOrder, in.IconKey, in.IsActive)
}