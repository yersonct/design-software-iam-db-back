package role

import (
	"context"

	domainrole "github.com/yersonct/iam-service/internal/domain/role"
)

type CreateRoleInput struct {
	Name        string
	DisplayName string
	Description string
}

type CreateRoleUseCase struct {
	repo domainrole.Repository
}

func NewCreateRoleUseCase(repo domainrole.Repository) *CreateRoleUseCase {
	return &CreateRoleUseCase{repo: repo}
}

func (uc *CreateRoleUseCase) Execute(ctx context.Context, input CreateRoleInput) (*domainrole.Role, error) {
	exists, err := uc.repo.ExistsByName(ctx, input.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domainrole.ErrRoleNameAlreadyExists
	}

	r := &domainrole.Role{
		Name:        input.Name,
		DisplayName: input.DisplayName,
		Description: input.Description,
	}

	if err := uc.repo.Create(ctx, r); err != nil {
		return nil, err
	}

	return r, nil
}