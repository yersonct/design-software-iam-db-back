package role

import (
	"context"

	domainrole "github.com/yersonct/iam-service/internal/domain/role"
)

type DeleteRoleUseCase struct {
	repo domainrole.Repository
}

func NewDeleteRoleUseCase(repo domainrole.Repository) *DeleteRoleUseCase {
	return &DeleteRoleUseCase{repo: repo}
}

func (uc *DeleteRoleUseCase) Execute(ctx context.Context, id string) error {
	current, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Misma regla de dominio que en Update: un rol de sistema jamás
	// se elimina, sin importar si tiene usuarios asignados o no.
	if err := current.EnsureDeletable(); err != nil {
		return err
	}

	return uc.repo.Delete(ctx, id)
}