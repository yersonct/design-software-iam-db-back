package userrole

import "time"

type UserRole struct {
	ID               string
	UserID           string
	RoleID           string
	TrainingCenterID *string
	AssignedBy       string
	AssignedAt       time.Time
	ExpiresAt        *time.Time
}

func (ur *UserRole) IsExpired(now time.Time) bool {
	return ur.ExpiresAt != nil && ur.ExpiresAt.Before(now)
}

// UserRoleItem: fila enriquecida (join con role) para GET /users/{id}/roles.
type UserRoleItem struct {
	ID               string
	RoleID           string
	RoleName         string
	RoleDisplayName  string
	TrainingCenterID *string
	AssignedBy       string
	AssignedAt       time.Time
	ExpiresAt        *time.Time
}