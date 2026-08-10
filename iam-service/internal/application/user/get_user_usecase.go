package user

import (
	"context"

	domainrole "github.com/yersonct/iam-service/internal/domain/role"
	domainuser "github.com/yersonct/iam-service/internal/domain/user"
)

type GetUserOutput struct {
	User  *domainuser.User
	Roles []*domainrole.Role
}

type GetUserUseCase struct {
	userRepo domainuser.Repository
	roleRepo domainrole.Repository
}

func NewGetUserUseCase(userRepo domainuser.Repository, roleRepo domainrole.Repository) *GetUserUseCase {
	return &GetUserUseCase{userRepo: userRepo, roleRepo: roleRepo}
}

func (uc *GetUserUseCase) Execute(ctx context.Context, userID string) (*GetUserOutput, error) {
	u, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	roles, err := uc.roleRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &GetUserOutput{User: u, Roles: roles}, nil
}