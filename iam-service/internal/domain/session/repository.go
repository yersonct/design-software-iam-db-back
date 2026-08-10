// repository.go
package session

import "context"

type Repository interface {
	Create(ctx context.Context, s *Session) error
	FindByHash(ctx context.Context, hash string) (*Session, error)
	Revoke(ctx context.Context, id string) error
	RevokeAllByUserID(ctx context.Context, userID string) error
}
