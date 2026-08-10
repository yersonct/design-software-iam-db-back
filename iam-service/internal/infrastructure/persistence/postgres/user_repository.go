package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
	"github.com/yersonct/iam-service/internal/domain/user"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByEmail trae el usuario junto con TODOS sus roles activos y vigentes
// (no solo el más reciente). Antes esto hacía LEFT JOIN + ORDER BY
// assigned_at DESC NULLS LAST LIMIT 1, lo que descartaba silenciosamente
// cualquier rol adicional del usuario (ej. INSTRUCTOR + COORDINATOR).
// Usamos una subconsulta correlacionada con array_agg para mantener 1 fila
// por usuario sin el fan-out del JOIN, y COALESCE a array vacío cuando no
// tiene ningún rol vigente (mismo caso que antes daba '').
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	const query = `
        SELECT
            u.id,
            u.email,
            u.password_hash,
            u.is_active,
            u.failed_attempts,
            u.locked_until,
            u.last_login_at,
            COALESCE(
                (
                    SELECT array_agg(r.name ORDER BY r.name)
                    FROM rbac.user_role ur
                    JOIN rbac.role r ON r.id = ur.role_id
                    WHERE ur.user_id = u.id
                      AND (ur.expires_at IS NULL OR ur.expires_at > now())
                ),
                ARRAY[]::varchar[]
            ) AS role_names
        FROM identity.user u
        WHERE lower(u.email) = lower($1)
    `

	var u user.User
	var lockedUntil sql.NullTime
	var lastLoginAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.IsActive,
		&u.FailedAttempts,
		&lockedUntil,
		&lastLoginAt,
		pq.Array(&u.RoleNames),
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrNotFound
		}

		return nil, fmt.Errorf("find user by email: %w", err)
	}

	if lockedUntil.Valid {
		u.LockedUntil = &lockedUntil.Time
	}

	if lastLoginAt.Valid {
		u.LastLoginAt = &lastLoginAt.Time
	}

	return &u, nil
}

// FindByID: mismo criterio que FindByEmail (ver comentario ahí). Este
// método es el que usa RefreshTokenUseCase, así que es clave para el
// criterio de aceptación "un rol que expira a mitad de sesión deja de
// aparecer tras el próximo refresh" -- cada llamada vuelve a consultar
// rbac.user_role en vivo, nunca reusa roles calculados en el login.
func (r *UserRepository) FindByID(
	ctx context.Context,
	id string,
) (*user.User, error) {

	const query = `
		SELECT
			u.id,
			u.email,
			u.password_hash,
			u.first_name,
			u.last_name,
			u.actor_type,
			u.actor_id,
			u.is_active,
			u.failed_attempts,
			u.locked_until,
			u.last_login_at,
			u.created_at,
			COALESCE(
				(
					SELECT array_agg(r.name ORDER BY r.name)
					FROM rbac.user_role ur
					JOIN rbac.role r ON r.id = ur.role_id
					WHERE ur.user_id = u.id
					  AND (ur.expires_at IS NULL OR ur.expires_at > now())
				),
				ARRAY[]::varchar[]
			) AS role_names
		FROM identity.user u
		WHERE u.id = $1
	`

	var u user.User
	var actorID sql.NullString
	var lockedUntil sql.NullTime
	var lastLoginAt sql.NullTime

	err := r.db.QueryRowContext(
		ctx,
		query,
		id,
	).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.FirstName,
		&u.LastName,
		&u.ActorType,
		&actorID,
		&u.IsActive,
		&u.FailedAttempts,
		&lockedUntil,
		&lastLoginAt,
		&u.CreatedAt,
		pq.Array(&u.RoleNames),
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrNotFound
		}

		return nil, fmt.Errorf("find user by id: %w", err)
	}

	if actorID.Valid {
		id := actorID.String
		u.ActorID = &id
	}

	if lockedUntil.Valid {
		u.LockedUntil = &lockedUntil.Time
	}

	if lastLoginAt.Valid {
		u.LastLoginAt = &lastLoginAt.Time
	}

	return &u, nil
}

func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	const query = `
        UPDATE identity.user
        SET
            failed_attempts = $1,
            locked_until = $2,
            last_login_at = $3,
            updated_at = now()
        WHERE id = $4
    `

	result, err := r.db.ExecContext(
		ctx,
		query,
		u.FailedAttempts,
		u.LockedUntil,
		u.LastLoginAt,
		u.ID,
	)
	if err != nil {
		return fmt.Errorf("update user login state: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check updated user: %w", err)
	}

	if rowsAffected == 0 {
		return user.ErrNotFound
	}

	return nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, userID string, passwordHash string) error {
	const query = `
                UPDATE identity.user
                SET
                        password_hash = $1,
                        updated_at = now()
                WHERE id = $2
        `

	result, err := r.db.ExecContext(ctx, query, passwordHash, userID)
	if err != nil {
		return fmt.Errorf("update user password: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check updated password: %w", err)
	}

	if rowsAffected == 0 {
		return user.ErrNotFound
	}

	return nil
}

// ExistsByEmail valida unicidad ANTES del insert, comparando en
// minúsculas igual que el índice único uq_user_email_lower.
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM identity.user WHERE lower(email) = lower($1))`

	var exists bool
	if err := r.db.QueryRowContext(ctx, query, email).Scan(&exists); err != nil {
		return false, fmt.Errorf("check email existence: %w", err)
	}

	return exists, nil
}

// Create inserta un nuevo usuario. Además del chequeo previo con
// ExistsByEmail, se captura el error de constraint única de Postgres
// (código 23505) como defensa extra ante condiciones de carrera.
func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	const query = `
                INSERT INTO identity.user (
                        email,
                        password_hash,
                        first_name,
                        last_name,
                        actor_type,
                        actor_id,
                        is_active
                ) VALUES ($1, $2, $3, $4, $5, $6, $7)
                RETURNING id
        `

	err := r.db.QueryRowContext(
		ctx,
		query,
		u.Email,
		u.PasswordHash,
		u.FirstName,
		u.LastName,
		string(u.ActorType),
		u.ActorID,
		u.IsActive,
	).Scan(&u.ID)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return user.ErrEmailAlreadyExists
		}

		return fmt.Errorf("create user: %w", err)
	}

	return nil
}
// List retorna una página de usuarios filtrados por actor_type y/o
// is_active. Usa COUNT(*) OVER() para traer el total en la misma
// consulta, sin necesitar un segundo roundtrip a la base de datos.
func (r *UserRepository) List(
	ctx context.Context,
	filter user.UserFilter,
	limit, offset int,
) ([]*user.User, int, error) {

	query := `
		SELECT
			u.id,
			u.email,
			u.first_name,
			u.last_name,
			u.actor_type,
			u.actor_id,
			u.is_active,
			u.failed_attempts,
			u.locked_until,
			u.last_login_at,
			u.created_at,
			COUNT(*) OVER() AS total_count
		FROM identity.user u
		WHERE 1 = 1
	`

	args := []interface{}{}
	argIdx := 1

	if filter.ActorType != nil {
		query += fmt.Sprintf(" AND u.actor_type = $%d", argIdx)
		args = append(args, string(*filter.ActorType))
		argIdx++
	}

	if filter.IsActive != nil {
		query += fmt.Sprintf(" AND u.is_active = $%d", argIdx)
		args = append(args, *filter.IsActive)
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY u.created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []*user.User
	total := 0

	for rows.Next() {
		var u user.User
		var actorID sql.NullString
		var lockedUntil sql.NullTime
		var lastLoginAt sql.NullTime

		err := rows.Scan(
			&u.ID,
			&u.Email,
			&u.FirstName,
			&u.LastName,
			&u.ActorType,
			&actorID,
			&u.IsActive,
			&u.FailedAttempts,
			&lockedUntil,
			&lastLoginAt,
			&u.CreatedAt,
			&total,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan user row: %w", err)
		}

		if actorID.Valid {
			id := actorID.String
			u.ActorID = &id
		}

		if lockedUntil.Valid {
			u.LockedUntil = &lockedUntil.Time
		}

		if lastLoginAt.Valid {
			u.LastLoginAt = &lastLoginAt.Time
		}

		users = append(users, &u)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate user rows: %w", err)
	}

	return users, total, nil
}
// UpdateProfile edita nombre y correo. La unicidad de email se apoya en
// el constraint uq_user_email de la base de datos como defensa final,
// además de la validación previa que hace el caso de uso.
func (r *UserRepository) UpdateProfile(
	ctx context.Context,
	id string,
	email string,
	firstName string,
	lastName string,
) error {
	const query = `
		UPDATE identity.user
		SET
			email = $1,
			first_name = $2,
			last_name = $3,
			updated_at = now()
		WHERE id = $4
	`

	result, err := r.db.ExecContext(ctx, query, email, firstName, lastName, id)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return user.ErrEmailAlreadyExists
		}
		return fmt.Errorf("update user profile: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check updated profile: %w", err)
	}

	if rowsAffected == 0 {
		return user.ErrNotFound
	}

	return nil
}

// SetActive activa o desactiva una cuenta.
func (r *UserRepository) SetActive(ctx context.Context, id string, isActive bool) error {
	const query = `
		UPDATE identity.user
		SET
			is_active = $1,
			updated_at = now()
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, isActive, id)
	if err != nil {
		return fmt.Errorf("set user active status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check set active: %w", err)
	}

	if rowsAffected == 0 {
		return user.ErrNotFound
	}

	return nil
}

// Unlock limpia el bloqueo por intentos fallidos, permitiendo login
// inmediato si las credenciales son correctas.
func (r *UserRepository) Unlock(ctx context.Context, id string) error {
	const query = `
		UPDATE identity.user
		SET
			locked_until = NULL,
			failed_attempts = 0,
			updated_at = now()
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("unlock user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check unlock: %w", err)
	}

	if rowsAffected == 0 {
		return user.ErrNotFound
	}

	return nil
}