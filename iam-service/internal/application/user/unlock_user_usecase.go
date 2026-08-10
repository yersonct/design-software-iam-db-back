package user

import (
	"context"

	domainuser "github.com/yersonct/iam-service/internal/domain/user"
)

type UnlockUserUseCase struct {
	userRepo domainuser.Repository
}

func NewUnlockUserUseCase(userRepo domainuser.Repository) *UnlockUserUseCase {
	return &UnlockUserUseCase{userRepo: userRepo}
}

func (uc *UnlockUserUseCase) Execute(ctx context.Context, userID string) error {
	return uc.userRepo.Unlock(ctx, userID)
}