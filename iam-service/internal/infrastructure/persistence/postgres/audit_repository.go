package postgres

import (
	"context"
	"database/sql"
	"strings"
	"fmt"
	domainaudit "github.com/yersonct/iam-service/internal/domain/audit"
)

type AuditRepository struct {
	db *sql.DB
}

func NewAuditRepository(db *sql.DB) *AuditRepository {
	return &AuditRepository{
		db: db,
	}
}

// LogLoginAttempt registra un intento de inicio de sesión.
func (r *AuditRepository) LogLoginAttempt(
	ctx context.Context,
	email string,
	success bool,
	reason string,
) {
	outcome := mapLoginOutcome(success, reason)

	const query = `
		INSERT INTO identity_audit.audit_login (
			email_attempted,
			outcome,
			attempted_at
		)
		VALUES ($1, $2, now())
	`

	if _, err := r.db.ExecContext(
		ctx,
		query,
		email,
		outcome,
	); err != nil {
		// La auditoría no debe romper el login.
		// Registramos el error para diagnóstico.
		fmt.Printf("error registrando intento de login: %v\n", err)
	}
}

// Log registra un evento de auditoría.
func (r *AuditRepository) Log(
	ctx context.Context,
	userID string,
	action string,
) error {
	// TODO: implementar otros eventos de auditoría.
	return nil
}

func mapLoginOutcome(success bool, reason string) string {
	if success {
		return "SUCCESS"
	}

	switch reason {
	case "user_not_found":
		return "USER_NOT_FOUND"

	case "bad_password":
		return "INVALID_PASSWORD"

	case "locked":
		return "ACCOUNT_LOCKED"

	default:
		return "INVALID_PASSWORD"
	}
}

// ListLoginAttempts implementa domainaudit.Repository. Solo lectura --
// la tabla identity_audit.audit_login es append-only, no se toca aquí.
func (r *AuditRepository) ListLoginAttempts(ctx context.Context, filter domainaudit.ListFilter) (*domainaudit.ListResult, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	argN := 1

	if filter.Email != nil && *filter.Email != "" {
		where = append(where, fmt.Sprintf("email_attempted ILIKE $%d", argN))
		args = append(args, "%"+*filter.Email+"%")
		argN++
	}
	if filter.Success != nil {
		if *filter.Success {
			where = append(where, "outcome = 'SUCCESS'")
		} else {
			where = append(where, "outcome != 'SUCCESS'")
		}
	}
	if filter.From != nil {
		where = append(where, fmt.Sprintf("attempted_at >= $%d", argN))
		args = append(args, *filter.From)
		argN++
	}
	if filter.To != nil {
		where = append(where, fmt.Sprintf("attempted_at <= $%d", argN))
		args = append(args, *filter.To)
		argN++
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	countQuery := "SELECT COUNT(*) FROM identity_audit.audit_login WHERE " + whereClause
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count login attempts: %w", err)
	}

	offset := (filter.Page - 1) * filter.PageSize
	query := fmt.Sprintf(`
		SELECT id, email_attempted, outcome, attempted_at
		FROM identity_audit.audit_login
		WHERE %s
		ORDER BY attempted_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argN, argN+1)
	args = append(args, filter.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list login attempts: %w", err)
	}
	defer rows.Close()

	items := make([]domainaudit.LoginAttempt, 0, filter.PageSize)
	for rows.Next() {
		var it domainaudit.LoginAttempt
		if err := rows.Scan(&it.ID, &it.EmailAttempted, &it.Outcome, &it.AttemptedAt); err != nil {
			return nil, fmt.Errorf("scan login attempt row: %w", err)
		}
		items = append(items, it)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate login attempts: %w", err)
	}

	return &domainaudit.ListResult{
		Items: items, Total: total, Page: filter.Page, PageSize: filter.PageSize,
	}, nil
}