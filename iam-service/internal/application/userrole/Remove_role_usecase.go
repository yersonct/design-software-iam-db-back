package userrole

import (
	"context"

	domainuserrole "github.com/yersonct/iam-service/internal/domain/userrole"
)

type RemoveRoleUseCase struct {
	repo domainuserrole.Repository
}

func NewRemoveRoleUseCase(repo domainuserrole.Repository) *RemoveRoleUseCase {
	return &RemoveRoleUseCase{repo: repo}
}

func (uc *RemoveRoleUseCase) Execute(ctx context.Context, userID, roleID string, trainingCenterID *string) error {
	return uc.repo.Remove(ctx, userID, roleID, trainingCenterID)
}