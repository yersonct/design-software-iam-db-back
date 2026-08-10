package passwordreset

import "context"

type Repository interface {
	Create(ctx context.Context, r *PasswordResetRequest) error
	FindByHash(ctx context.Context, hash string) (*PasswordResetRequest, error)
	MarkUsed(ctx context.Context, id string) error
}