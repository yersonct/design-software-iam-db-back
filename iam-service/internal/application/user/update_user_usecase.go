package user

import (
	"context"
	"fmt"
"strings"


	domainuser "github.com/yersonct/iam-service/internal/domain/user"
)

type UpdateUserInput struct {
	ID        string
	Email     string
	FirstName string
	LastName  string
}

type UpdateUserUseCase struct {
	userRepo domainuser.Repository
}

func NewUpdateUserUseCase(userRepo domainuser.Repository) *UpdateUserUseCase {
	return &UpdateUserUseCase{userRepo: userRepo}
}

func (uc *UpdateUserUseCase) Execute(ctx context.Context, in UpdateUserInput) error {
	current, err := uc.userRepo.FindByID(ctx, in.ID)
	if err != nil {
		return err
	}

	// Solo validamos unicidad si el correo realmente cambió. Si no cambió,
	// que exista un registro con ese email es esperado: es el mismo usuario.
// ...
	emailChanged := !strings.EqualFold(current.Email, in.Email)

	if emailChanged {
		exists, err := uc.userRepo.ExistsByEmail(ctx, in.Email)
		if err != nil {
			return fmt.Errorf("check email existence: %w", err)
		}
		if exists {
			return domainuser.ErrEmailAlreadyExists
		}
	}

	return uc.userRepo.UpdateProfile(ctx, in.ID, in.Email, in.FirstName, in.LastName)
}

