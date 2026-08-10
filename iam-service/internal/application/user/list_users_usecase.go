package user

import (
	"context"

	domainuser "github.com/yersonct/iam-service/internal/domain/user"
)

type ListUsersInput struct {
	Page      int
	PageSize  int
	ActorType *domainuser.ActorType
	IsActive  *bool
}

type ListUsersOutput struct {
	Users      []*domainuser.User
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}

type ListUsersUseCase struct {
	userRepo domainuser.Repository
}

func NewListUsersUseCase(userRepo domainuser.Repository) *ListUsersUseCase {
	return &ListUsersUseCase{userRepo: userRepo}
}

func (uc *ListUsersUseCase) Execute(ctx context.Context, in ListUsersInput) (*ListUsersOutput, error) {
	page := in.Page
	if page < 1 {
		page = 1
	}

	pageSize := in.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	filter := domainuser.UserFilter{
		ActorType: in.ActorType,
		IsActive:  in.IsActive,
	}

	users, total, err := uc.userRepo.List(ctx, filter, pageSize, offset)
	if err != nil {
		return nil, err
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	return &ListUsersOutput{
		Users:      users,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}