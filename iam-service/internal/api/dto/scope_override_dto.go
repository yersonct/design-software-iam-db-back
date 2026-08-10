package dto

import "time"

type CreateScopeOverrideRequest struct {
	FeatureID string     `json:"feature_id" binding:"required,uuid"`
	ScopeType string     `json:"scope_type" binding:"required"`
	IsAllowed bool       `json:"is_allowed"`
	Reason    string     `json:"reason" binding:"required"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type ScopeOverrideResponse struct {
	ID                 string     `json:"id"`
	FeatureID          string     `json:"feature_id"`
	FeatureName        string     `json:"feature_name"`
	FeatureDisplayName string     `json:"feature_display_name"`
	ScopeType          string     `json:"scope_type"`
	IsAllowed          bool       `json:"is_allowed"`
	Reason             string     `json:"reason"`
	GrantedBy          string     `json:"granted_by"`
	CreatedAt          time.Time  `json:"created_at"`
	ExpiresAt          *time.Time `json:"expires_at,omitempty"`
	IsExpired          bool       `json:"is_expired"`
}