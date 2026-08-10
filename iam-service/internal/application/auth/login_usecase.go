package auth

import (
	"context"
	"errors"
	"time"

	"github.com/yersonct/iam-service/internal/domain/session"
	"github.com/yersonct/iam-service/internal/domain/user"
)

const (
	maxFailedAttempts = 5
	lockDuration      = 15 * time.Minute
)

// Puertos que el caso de uso necesita — definidos como interfaces AQUÍ
// para no acoplarse a implementaciones concretas de infra
type PasswordHasher interface {
	Compare(hash, plain string) bool
	Hash(plain string) (string, error)
}

type TokenGenerator interface {
	GenerateAccessToken(userID string, roles []string) (string, error)
	GenerateRefreshToken() (plain string, hash string, err error)
	GenerateResetToken() (plain string, hash string, err error)
}

type AuditLogger interface {
	LogLoginAttempt(ctx context.Context, email string, success bool, reason string)
}

type LoginUseCase struct {
	userRepo    user.Repository
	sessionRepo session.Repository
	hasher      PasswordHasher
	tokens      TokenGenerator
	audit       AuditLogger
	clock       func() time.Time
}

func NewLoginUseCase(
	userRepo user.Repository,
	sessionRepo session.Repository,
	hasher PasswordHasher,
	tokens TokenGenerator,
	audit AuditLogger,
) *LoginUseCase {
	return &LoginUseCase{
		userRepo: userRepo, sessionRepo: sessionRepo,
		hasher: hasher, tokens: tokens, audit: audit,
		clock: time.Now,
	}
}

func (uc *LoginUseCase) Execute(ctx context.Context, in LoginInput) (*LoginOutput, error) {
	now := uc.clock()

	u, err := uc.userRepo.FindByEmail(ctx, in.Email)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			uc.audit.LogLoginAttempt(ctx, in.Email, false, "user_not_found")
			return nil, user.ErrInvalidCredentials
		}

		return nil, err
	}

	if !u.IsActive {
		uc.audit.LogLoginAttempt(ctx, in.Email, false, "inactive")
		return nil, user.ErrAccountInactive // criterio #3: rechazo explícito
	}

	if u.IsLocked(now) {
		uc.audit.LogLoginAttempt(ctx, in.Email, false, "locked")
		return nil, user.ErrAccountLocked // criterio #2
	}

	if !uc.hasher.Compare(u.PasswordHash, in.Password) {
		u.RegisterFailedAttempt(maxFailedAttempts, lockDuration, now)
		_ = uc.userRepo.Update(ctx, u)
		uc.audit.LogLoginAttempt(ctx, in.Email, false, "bad_password")
		if u.IsLocked(now) {
			return nil, user.ErrAccountLocked
		}
		return nil, user.ErrInvalidCredentials
	}

	// Éxito
	u.RegisterSuccessfulLogin(now)
	_ = uc.userRepo.Update(ctx, u)

	access, err := uc.tokens.GenerateAccessToken(u.ID, u.RoleNames)
	if err != nil {
		return nil, err
	}
	refreshPlain, refreshHash, err := uc.tokens.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	_ = uc.sessionRepo.Create(ctx, &session.Session{
		UserID:           u.ID,
		RefreshTokenHash: refreshHash, // solo el hash se persiste
		ExpiresAt:        now.Add(7 * 24 * time.Hour),
		CreatedAt:        now,
	})

	uc.audit.LogLoginAttempt(ctx, in.Email, true, "success")

	return &LoginOutput{
		AccessToken:  access,
		RefreshToken: refreshPlain, // solo sale al cliente aquí, nunca se guarda así
		Roles:        u.RoleNames,
		UserID:       u.ID,
		// u.RoleNames viene de FindByEmail: array_agg sobre rbac.user_role
		// filtrado por (expires_at IS NULL OR expires_at > now()), con
		// TODOS los roles vigentes, no solo el más reciente. Si no tiene
		// ningún rol activo, llega como slice vacío -- ese es el mismo
		// criterio que usamos aquí, sin consulta extra.
		HasActiveRole: len(u.RoleNames) > 0,
	}, nil
}