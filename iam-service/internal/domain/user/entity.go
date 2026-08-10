package user

import "time"

// ActorType refleja el CHECK constraint ck_user_actor_type de la tabla
// identity.user: solo estos tres valores son válidos.
type ActorType string

const (
	ActorTypeUser       ActorType = "USER"
	ActorTypeInstructor ActorType = "INSTRUCTOR"
	ActorTypeLearner    ActorType = "LEARNER"
)

func (a ActorType) IsValid() bool {
	switch a {
	case ActorTypeUser, ActorTypeInstructor, ActorTypeLearner:
		return true
	default:
		return false
	}
}

type User struct {
	ID             string
	Email          string
	PasswordHash   string
	FirstName      string
	LastName       string
	ActorType      ActorType
	ActorID        *string // nullable en BD: solo aplica si el actor vive en actors-service
	IsActive       bool
	FailedAttempts int
	LockedUntil    *time.Time
	LastLoginAt    *time.Time
	CreatedAt      time.Time
	RoleNames []string
}

// Reglas de negocio VIVEN en la entidad, no en el caso de uso ni el handler
func (u *User) IsLocked(now time.Time) bool {
	return u.LockedUntil != nil && u.LockedUntil.After(now)
}

func (u *User) RegisterFailedAttempt(maxAttempts int, lockDuration time.Duration, now time.Time) {
	u.FailedAttempts++
	if u.FailedAttempts >= maxAttempts {
		locked := now.Add(lockDuration)
		u.LockedUntil = &locked
	}
}

func (u *User) RegisterSuccessfulLogin(now time.Time) {
	u.FailedAttempts = 0
	u.LockedUntil = nil
	u.LastLoginAt = &now
}