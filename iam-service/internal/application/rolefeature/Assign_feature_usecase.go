package rolefeature

import (
	"context"

	domainrolefeature "github.com/yersonct/iam-service/internal/domain/rolefeature"
)

type AssignFeatureInput struct {
	RoleID    string
	FeatureID string
	ScopeType string
}

type AssignFeatureUseCase struct {
	repo domainrolefeature.Repository
}

func NewAssignFeatureUseCase(repo domainrolefeature.Repository) *AssignFeatureUseCase {
	return &AssignFeatureUseCase{repo: repo}
}

func (uc *AssignFeatureUseCase) Execute(ctx context.Context, input AssignFeatureInput) (*domainrolefeature.RoleFeature, error) {
	scope := domainrolefeature.ScopeType(input.ScopeType)

	// Validación en el dominio, no solo en el binding:"oneof" del DTO --
	// mismo patrón que domain/catalog.ActionLevel.
	if !scope.IsValid() {
		return nil, domainrolefeature.ErrInvalidScopeType
	}

	rf := &domainrolefeature.RoleFeature{
		RoleID:    input.RoleID,
		FeatureID: input.FeatureID,
		ScopeType: scope,
	}

	if err := uc.repo.Assign(ctx, rf); err != nil {
		return nil, err
	}

	return rf, nil
}