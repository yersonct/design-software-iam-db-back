package user

import (
	"context"
	"errors"

	domainuser "github.com/yersonct/iam-service/internal/domain/user"
)

var ErrCannotDeactivateSelf = errors.New("cannot_deactivate_self")

type SetUserStatusInput struct {
	TargetUserID    string
	RequestingUserID string
	IsActive        bool
}

type SetUserStatusUseCase struct {
	userRepo domainuser.Repository
}

func NewSetUserStatusUseCase(userRepo domainuser.Repository) *SetUserStatusUseCase {
	return &SetUserStatusUseCase{userRepo: userRepo}
}

func (uc *SetUserStatusUseCase) Execute(ctx context.Context, in SetUserStatusInput) error {
	return uc.userRepo.SetActive(ctx, in.TargetUserID, in.IsActive)
}