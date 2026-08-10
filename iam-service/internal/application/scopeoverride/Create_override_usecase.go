package scopeoverride

import (
	"context"
	"strings"
	"time"

	domainrolefeature "github.com/yersonct/iam-service/internal/domain/rolefeature"
	domainscopeoverride "github.com/yersonct/iam-service/internal/domain/scopeoverride"
)

type CreateOverrideInput struct {
	UserID     string
	FeatureID  string
	ScopeType  string
	IsAllowed  bool
	Reason     string
	GrantedBy  string
	ExpiresAt  *time.Time
}

type CreateOverrideUseCase struct {
	repo domainscopeoverride.Repository
}

func NewCreateOverrideUseCase(repo domainscopeoverride.Repository) *CreateOverrideUseCase {
	return &CreateOverrideUseCase{repo: repo}
}

func (uc *CreateOverrideUseCase) Execute(ctx context.Context, input CreateOverrideInput) (*domainscopeoverride.ScopeOverride, error) {
	// Criterio de aceptación: "reason obligatorio" con mensaje claro, no 500.
	if strings.TrimSpace(input.Reason) == "" {
		return nil, domainscopeoverride.ErrReasonRequired
	}
	if !domainrolefeature.ScopeType(input.ScopeType).IsValid() {
		return nil, domainscopeoverride.ErrInvalidScopeType
	}
	if input.ExpiresAt != nil && !input.ExpiresAt.After(time.Now()) {
		return nil, domainscopeoverride.ErrInvalidExpiresAt
	}

	o := &domainscopeoverride.ScopeOverride{
		UserID:    input.UserID,
		FeatureID: input.FeatureID,
		ScopeType: input.ScopeType,
		IsAllowed: input.IsAllowed,
		Reason:    strings.TrimSpace(input.Reason),
		GrantedBy: input.GrantedBy,
		ExpiresAt: input.ExpiresAt,
	}

	if err := uc.repo.Create(ctx, o); err != nil {
		return nil, err
	}
	return o, nil
}