package catalog

import (
	"context"
	"fmt"

	domaincatalog "github.com/yersonct/iam-service/internal/domain/catalog"
)

type CreateModuleInput struct {
	Code         string
	Name         string
	Description  *string
	DisplayOrder int16
	IconKey      *string
}

type CreateModuleUseCase struct {
	moduleRepo domaincatalog.ModuleRepository
}

func NewCreateModuleUseCase(moduleRepo domaincatalog.ModuleRepository) *CreateModuleUseCase {
	return &CreateModuleUseCase{moduleRepo: moduleRepo}
}

func (uc *CreateModuleUseCase) Execute(ctx context.Context, in CreateModuleInput) (*domaincatalog.Module, error) {
	exists, err := uc.moduleRepo.ExistsByCode(ctx, in.Code)
	if err != nil {
		return nil, fmt.Errorf("check module code existence: %w", err)
	}
	if exists {
		return nil, domaincatalog.ErrModuleCodeAlreadyExists
	}

	m := &domaincatalog.Module{
		Code:         in.Code,
		Name:         in.Name,
		Description:  in.Description,
		DisplayOrder: in.DisplayOrder,
		IconKey:      in.IconKey,
	}

	if err := uc.moduleRepo.Create(ctx, m); err != nil {
		return nil, err
	}

	return m, nil
}