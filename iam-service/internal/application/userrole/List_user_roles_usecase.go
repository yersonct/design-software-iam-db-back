package userrole

import (
	"context"

	domainuserrole "github.com/yersonct/iam-service/internal/domain/userrole"
)

type ListUserRolesUseCase struct {
	repo domainuserrole.Repository
}

func NewListUserRolesUseCase(repo domainuserrole.Repository) *ListUserRolesUseCase {
	return &ListUserRolesUseCase{repo: repo}
}

func (uc *ListUserRolesUseCase) Execute(ctx context.Context, userID string) ([]domainuserrole.UserRoleItem, error) {
	return uc.repo.ListByUser(ctx, userID)
}