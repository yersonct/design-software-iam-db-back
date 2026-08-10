package scopeoverride

import "time"

type ScopeOverride struct {
	ID         string
	UserID     string
	FeatureID  string
	ScopeType  string
	IsAllowed  bool
	Reason     string
	GrantedBy  string
	CreatedAt  time.Time
	ExpiresAt  *time.Time
}

func (o *ScopeOverride) IsExpired(now time.Time) bool {
	return o.ExpiresAt != nil && o.ExpiresAt.Before(now)
}

// ScopeOverrideItem: fila enriquecida (join con feature) para GET /users/{id}/scope-overrides.
type ScopeOverrideItem struct {
	ID                 string
	FeatureID          string
	FeatureName        string
	FeatureDisplayName string
	ScopeType          string
	IsAllowed          bool
	Reason             string
	GrantedBy          string
	CreatedAt          time.Time
	ExpiresAt          *time.Time
}