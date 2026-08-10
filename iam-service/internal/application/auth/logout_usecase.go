package auth

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"

	"github.com/yersonct/iam-service/internal/domain/session"
)

type LogoutUseCase struct {
	sessionRepo session.Repository
}

func NewLogoutUseCase(sessionRepo session.Repository) *LogoutUseCase {
	return &LogoutUseCase{
		sessionRepo: sessionRepo,
	}
}

// Execute revoca la sesión asociada al refresh token recibido.
// Es idempotente a propósito: si el token no existe o ya estaba
// revocado, no se considera un error — el resultado que le importa
// al usuario (quedar deslogueado) igual se cumple.
func (uc *LogoutUseCase) Execute(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}

	sum := sha256.Sum256([]byte(refreshToken))
	tokenHash := hex.EncodeToString(sum[:])

	s, err := uc.sessionRepo.FindByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	if s.IsRevoked {
		return nil
	}

	if err := uc.sessionRepo.Revoke(ctx, s.ID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	return nil
}