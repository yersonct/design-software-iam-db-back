package dto

import "time"

type AssignRoleRequest struct {
	RoleID           string     `json:"role_id" binding:"required,uuid"`
	TrainingCenterID *string    `json:"training_center_id,omitempty" binding:"omitempty,uuid"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
}

type UserRoleResponse struct {
	ID               string     `json:"id"`
	RoleID           string     `json:"role_id"`
	RoleName         string     `json:"role_name"`
	RoleDisplayName  string     `json:"role_display_name"`
	TrainingCenterID *string    `json:"training_center_id,omitempty"`
	AssignedBy       string     `json:"assigned_by"`
	AssignedAt       time.Time  `json:"assigned_at"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
	IsExpired        bool       `json:"is_expired"`
}