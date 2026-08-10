package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/yersonct/iam-service/internal/domain/session"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{
		db: db,
	}
}

// Create guarda únicamente el hash del refresh token.
// El token plano nunca se almacena en PostgreSQL.
func (r *SessionRepository) Create(ctx context.Context, s *session.Session) error {
	const query = `
		INSERT INTO session.refresh_token (
			user_id,
			token_hash,
			expires_at,
			is_revoked,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		s.UserID,
		s.RefreshTokenHash,
		s.ExpiresAt,
		s.IsRevoked,
		s.CreatedAt,
	).Scan(&s.ID)

	if err != nil {
		return fmt.Errorf("create refresh token session: %w", err)
	}

	return nil
}

// FindByHash busca una sesión por el hash del refresh token.
func (r *SessionRepository) FindByHash(
	ctx context.Context,
	hash string,
) (*session.Session, error) {
	const query = `
		SELECT
			id,
			user_id,
			token_hash,
			expires_at,
			is_revoked,
			created_at
		FROM session.refresh_token
		WHERE token_hash = $1
	`

	var s session.Session

	err := r.db.QueryRowContext(
		ctx,
		query,
		hash,
	).Scan(
		&s.ID,
		&s.UserID,
		&s.RefreshTokenHash,
		&s.ExpiresAt,
		&s.IsRevoked,
		&s.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}

		return nil, fmt.Errorf("find refresh token session: %w", err)
	}

	return &s, nil
}

// Revoke invalida una sesión de refresh token.
func (r *SessionRepository) Revoke(
	ctx context.Context,
	id string,
) error {
	const query = `
		UPDATE session.refresh_token
		SET
			is_revoked = TRUE,
			revoked_at = now()
		WHERE id = $1
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		id,
	)

	if err != nil {
		return fmt.Errorf("revoke refresh token session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check revoked refresh token: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
// RevokeAllByUserID invalida todas las sesiones activas de un usuario.
// Se usa tras un reset-password exitoso, para forzar el cierre de
// cualquier sesión que pudiera seguir abierta en otros dispositivos.
func (r *SessionRepository) RevokeAllByUserID(ctx context.Context, userID string) error {
	const query = `
		UPDATE session.refresh_token
		SET
			is_revoked = TRUE,
			revoked_at = now()
		WHERE user_id = $1 AND is_revoked = FALSE
	`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("revoke all refresh tokens: %w", err)
	}

	return nil
}
