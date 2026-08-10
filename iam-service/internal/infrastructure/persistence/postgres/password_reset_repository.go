package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/yersonct/iam-service/internal/domain/passwordreset"
)

type PasswordResetRepository struct {
	db *sql.DB
}

func NewPasswordResetRepository(db *sql.DB) *PasswordResetRepository {
	return &PasswordResetRepository{db: db}
}

func (r *PasswordResetRepository) Create(ctx context.Context, req *passwordreset.PasswordResetRequest) error {
	const query = `
		INSERT INTO session.password_reset_request (
			user_id, token_hash, expires_at, is_used, requested_at, ip_address
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err := r.db.QueryRowContext(
		ctx, query,
		req.UserID, req.TokenHash, req.ExpiresAt, req.IsUsed, req.RequestedAt, req.IPAddress,
	).Scan(&req.ID)

	if err != nil {
		return fmt.Errorf("create password reset request: %w", err)
	}

	return nil
}

func (r *PasswordResetRepository) FindByHash(ctx context.Context, hash string) (*passwordreset.PasswordResetRequest, error) {
	const query = `
		SELECT id, user_id, token_hash, expires_at, is_used, requested_at, ip_address
		FROM session.password_reset_request
		WHERE token_hash = $1
		ORDER BY requested_at DESC
		LIMIT 1
	`

	var req passwordreset.PasswordResetRequest
	var ip sql.NullString

	err := r.db.QueryRowContext(ctx, query, hash).Scan(
		&req.ID, &req.UserID, &req.TokenHash, &req.ExpiresAt, &req.IsUsed, &req.RequestedAt, &ip,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("find password reset request: %w", err)
	}

	req.IPAddress = ip.String

	return &req, nil
}

func (r *PasswordResetRepository) MarkUsed(ctx context.Context, id string) error {
	const query = `
		UPDATE session.password_reset_request
		SET is_used = TRUE
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("mark password reset used: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check mark used rows: %w", err)
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}