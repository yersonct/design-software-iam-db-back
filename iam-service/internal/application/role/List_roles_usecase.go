package role

import (
	"context"

	domainrole "github.com/yersonct/iam-service/internal/domain/role"
)

type ListRolesUseCase struct {
	repo domainrole.Repository
}

func NewListRolesUseCase(repo domainrole.Repository) *ListRolesUseCase {
	return &ListRolesUseCase{repo: repo}
}

func (uc *ListRolesUseCase) Execute(ctx context.Context) ([]domainrole.Role, error) {
	return uc.repo.List(ctx)
}