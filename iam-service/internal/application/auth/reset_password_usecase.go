package auth

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"github.com/yersonct/iam-service/internal/domain/passwordreset"
	"github.com/yersonct/iam-service/internal/domain/session"
	"github.com/yersonct/iam-service/internal/domain/user"
)

var (
	ErrInvalidResetToken = errors.New("invalid_reset_token")
	ErrResetTokenExpired = errors.New("reset_token_expired")
	ErrResetTokenUsed    = errors.New("reset_token_used")
	ErrWeakPassword      = errors.New("weak_password")
)

type ResetPasswordUseCase struct {
	passwordResetRepo passwordreset.Repository
	userRepo          user.Repository
	sessionRepo       session.Repository
	hasher            PasswordHasher
	clock             func() time.Time
}

func NewResetPasswordUseCase(
	passwordResetRepo passwordreset.Repository,
	userRepo user.Repository,
	sessionRepo session.Repository,
	hasher PasswordHasher,
) *ResetPasswordUseCase {
	return &ResetPasswordUseCase{
		passwordResetRepo: passwordResetRepo,
		userRepo:          userRepo,
		sessionRepo:       sessionRepo,
		hasher:            hasher,
		clock:             time.Now,
	}
}

func (uc *ResetPasswordUseCase) Execute(ctx context.Context, token, newPassword string) error {
	if len(newPassword) < 8 {
		return ErrWeakPassword
	}

	sum := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(sum[:])

	req, err := uc.passwordResetRepo.FindByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidResetToken
		}
		return err
	}

	if req.IsUsed {
		return ErrResetTokenUsed
	}

	if !req.ExpiresAt.After(uc.clock()) {
		return ErrResetTokenExpired
	}

	u, err := uc.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return ErrInvalidResetToken
	}

	newHash, err := uc.hasher.Hash(newPassword)
	if err != nil {
		return err
	}

	if err := uc.userRepo.UpdatePassword(ctx, u.ID, newHash); err != nil {
		return err
	}

	if err := uc.passwordResetRepo.MarkUsed(ctx, req.ID); err != nil {
		return err
	}

	// Criterio de aceptación: tras el reset, se revocan automáticamente
	// todas las sesiones activas del usuario en cualquier dispositivo.
	return uc.sessionRepo.RevokeAllByUserID(ctx, u.ID)
}