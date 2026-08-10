package catalog

import (
	"context"
	"fmt"

	domaincatalog "github.com/yersonct/iam-service/internal/domain/catalog"
)

type CreateFeatureInput struct {
	ModuleID    string
	Code        string
	Name        string
	Description *string
	ActionLevel domaincatalog.ActionLevel
}

type CreateFeatureUseCase struct {
	featureRepo domaincatalog.FeatureRepository
}

func NewCreateFeatureUseCase(featureRepo domaincatalog.FeatureRepository) *CreateFeatureUseCase {
	return &CreateFeatureUseCase{featureRepo: featureRepo}
}

func (uc *CreateFeatureUseCase) Execute(ctx context.Context, in CreateFeatureInput) (*domaincatalog.Feature, error) {
	// Defensa en profundidad: el DTO ya valida esto con `oneof` antes de llegar
	// aquí, pero el caso de uso no debe confiar ciegamente en el transporte HTTP.
	if !in.ActionLevel.Valid() {
		return nil, domaincatalog.ErrInvalidActionLevel
	}

	moduleExists, err := uc.featureRepo.ModuleExists(ctx, in.ModuleID)
	if err != nil {
		return nil, fmt.Errorf("check module existence: %w", err)
	}
	if !moduleExists {
		return nil, domaincatalog.ErrModuleNotFound
	}

	codeExists, err := uc.featureRepo.ExistsByCode(ctx, in.Code)
	if err != nil {
		return nil, fmt.Errorf("check feature code existence: %w", err)
	}
	if codeExists {
		return nil, domaincatalog.ErrFeatureCodeAlreadyExists
	}

	f := &domaincatalog.Feature{
		ModuleID:    in.ModuleID,
		Code:        in.Code,
		Name:        in.Name,
		Description: in.Description,
		ActionLevel: in.ActionLevel,
	}

	if err := uc.featureRepo.Create(ctx, f); err != nil {
		return nil, err
	}

	return f, nil
}