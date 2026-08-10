package auth

import (
    "context"
    "crypto/sha256"
    "encoding/hex"
    "errors"
    "time"

    "github.com/yersonct/iam-service/internal/domain/session"
    "github.com/yersonct/iam-service/internal/domain/user"
)

type RefreshTokenUseCase struct {
    sessionRepo session.Repository
    userRepo    user.Repository
    tokens      TokenGenerator
    clock       func() time.Time
}

func NewRefreshTokenUseCase(
    sessionRepo session.Repository,
    userRepo user.Repository,
    tokens TokenGenerator,
) *RefreshTokenUseCase {
    return &RefreshTokenUseCase{
        sessionRepo: sessionRepo,
        userRepo:    userRepo,
        tokens:      tokens,
        clock:       time.Now,
    }
}

// Execute renueva el access token. Importante: u viene de FindByID, que en
// CADA llamada vuelve a consultar rbac.user_role filtrado por vigencia
// (expires_at IS NULL OR expires_at > now()). No reusamos los roles del
// login original ni los del token viejo -- así, si un rol expiró mientras
// la sesión seguía activa, el próximo refresh ya no lo incluye, sin que el
// usuario tenga que cerrar sesión y volver a entrar.
func (uc *RefreshTokenUseCase) Execute(
    ctx context.Context,
    refreshToken string,
) (string, []string, error) {

    if refreshToken == "" {
        return "", nil, errors.New("refresh token required")
    }

    sum := sha256.Sum256([]byte(refreshToken))
    tokenHash := hex.EncodeToString(sum[:])

    s, err := uc.sessionRepo.FindByHash(ctx, tokenHash)
    if err != nil {
        return "", nil, errors.New("invalid refresh token")
    }

    if s.IsRevoked {
        return "", nil, errors.New("refresh token revoked")
    }

    now := uc.clock()

    if !s.ExpiresAt.After(now) {
        return "", nil, errors.New("refresh token expired")
    }

    u, err := uc.userRepo.FindByID(ctx, s.UserID)
    if err != nil {
        return "", nil, errors.New("user not found")
    }

    if !u.IsActive {
        return "", nil, errors.New("account inactive")
    }

    accessToken, err := uc.tokens.GenerateAccessToken(
        u.ID,
        u.RoleNames,
    )
    if err != nil {
        return "", nil, err
    }

    return accessToken, u.RoleNames, nil
}