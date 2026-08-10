package user

import "errors"

var (
	ErrNotFound           = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountLocked      = errors.New("account locked")
	ErrAccountInactive    = errors.New("account inactive")

	// ErrEmailAlreadyExists se retorna cuando ya existe un usuario con ese
	// email (comparación case-insensitive, igual que el índice único
	// uq_user_email_lower en la base de datos).
	ErrEmailAlreadyExists = errors.New("email already exists")

	// ErrInvalidActorType se retorna cuando actor_type no es uno de los
	// valores permitidos por el CHECK constraint ck_user_actor_type.
	ErrInvalidActorType = errors.New("invalid actor type")
)