package role

import (
	"context"

	domainrole "github.com/yersonct/iam-service/internal/domain/role"
)

type UpdateRoleInput struct {
	ID          string
	DisplayName string
	Description string
}

type UpdateRoleUseCase struct {
	repo domainrole.Repository
}

func NewUpdateRoleUseCase(repo domainrole.Repository) *UpdateRoleUseCase {
	return &UpdateRoleUseCase{repo: repo}
}

func (uc *UpdateRoleUseCase) Execute(ctx context.Context, input UpdateRoleInput) error {
	// Buscamos el rol primero para poder aplicar la regla de dominio
	// "un rol de sistema no se puede editar" ANTES de tocar la base de
	// datos. Esto es lo que pide explícitamente la historia: "regla de
	// negocio a validar en el dominio, no solo en BD".
	current, err := uc.repo.FindByID(ctx, input.ID)
	if err != nil {
		return err
	}

	if err := current.EnsureEditable(); err != nil {
		return err
	}

	return uc.repo.Update(ctx, input.ID, input.DisplayName, input.Description)
}