package userrole

import (
	"context"
	"time"

	domaintc "github.com/yersonct/iam-service/internal/domain/trainingcenter"
	domainuserrole "github.com/yersonct/iam-service/internal/domain/userrole"
)

type AssignRoleInput struct {
	UserID           string
	RoleID           string
	TrainingCenterID *string
	AssignedBy       string
	ExpiresAt        *time.Time
}

type AssignRoleUseCase struct {
	repo   domainuserrole.Repository
	tcRepo domaintc.Repository
}

func NewAssignRoleUseCase(repo domainuserrole.Repository, tcRepo domaintc.Repository) *AssignRoleUseCase {
	return &AssignRoleUseCase{repo: repo, tcRepo: tcRepo}
}

func (uc *AssignRoleUseCase) Execute(ctx context.Context, input AssignRoleInput) (*domainuserrole.UserRole, error) {
	if input.TrainingCenterID != nil && !uc.tcRepo.Exists(*input.TrainingCenterID) {
		return nil, domainuserrole.ErrInvalidTrainingCenter
	}
	if input.ExpiresAt != nil && !input.ExpiresAt.After(time.Now()) {
		return nil, domainuserrole.ErrInvalidExpiresAt
	}

	ur := &domainuserrole.UserRole{
		UserID:           input.UserID,
		RoleID:           input.RoleID,
		TrainingCenterID: input.TrainingCenterID,
		AssignedBy:       input.AssignedBy,
		ExpiresAt:        input.ExpiresAt,
	}

	if err := uc.repo.Assign(ctx, ur); err != nil {
		return nil, err
	}
	return ur, nil
}