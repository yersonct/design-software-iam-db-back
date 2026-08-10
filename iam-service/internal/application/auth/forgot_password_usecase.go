package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/yersonct/iam-service/internal/domain/passwordreset"
	"github.com/yersonct/iam-service/internal/domain/user"
)

type EmailSender interface {
	SendPasswordResetEmail(ctx context.Context, to string, resetLink string) error
}

type ForgotPasswordUseCase struct {
	userRepo          user.Repository
	passwordResetRepo passwordreset.Repository
	tokens            TokenGenerator
	emailSender       EmailSender
	frontendResetURL  string
	clock             func() time.Time
}

func NewForgotPasswordUseCase(
	userRepo user.Repository,
	passwordResetRepo passwordreset.Repository,
	tokens TokenGenerator,
	emailSender EmailSender,
	frontendResetURL string,
) *ForgotPasswordUseCase {
	return &ForgotPasswordUseCase{
		userRepo:          userRepo,
		passwordResetRepo: passwordResetRepo,
		tokens:            tokens,
		emailSender:       emailSender,
		frontendResetURL:  frontendResetURL,
		clock:             time.Now,
	}
}

// Execute es intencionalmente "silencioso": nunca revela si el correo
// existe o no en el sistema (evita enumeración de usuarios). Solo
// devuelve error ante fallos reales de infraestructura (DB caída,
// SMTP caído, etc.), nunca por "usuario no encontrado".
func (uc *ForgotPasswordUseCase) Execute(ctx context.Context, email string) error {
	u, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return nil
		}
		return err
	}

	if !u.IsActive {
		return nil
	}

	plain, hash, err := uc.tokens.GenerateResetToken()
	if err != nil {
		return err
	}

	now := uc.clock()

	req := &passwordreset.PasswordResetRequest{
		UserID:      u.ID,
		TokenHash:   hash,
		ExpiresAt:   now.Add(1 * time.Hour),
		IsUsed:      false,
		RequestedAt: now,
	}

	if err := uc.passwordResetRepo.Create(ctx, req); err != nil {
		return err
	}

	resetLink := fmt.Sprintf("%s?token=%s", uc.frontendResetURL, plain)

	return uc.emailSender.SendPasswordResetEmail(ctx, u.Email, resetLink)
}