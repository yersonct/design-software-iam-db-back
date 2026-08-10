package authz

import (
	"context"
	"time"

	domainscopeoverride "github.com/yersonct/iam-service/internal/domain/scopeoverride"
)

// RoleFeatureChecker: ⚠️ aún necesito Role_feature_repository.go para
// implementar el cuerpo real de esto (paso 3, pendiente).
type RoleFeatureChecker interface {
	HasFeatureViaRole(ctx context.Context, userID, featureID, scopeType string) (bool, error)
}

type CheckPermissionUseCase struct {
	overrideRepo domainscopeoverride.Repository
	roleChecker  RoleFeatureChecker
}

func NewCheckPermissionUseCase(overrideRepo domainscopeoverride.Repository, roleChecker RoleFeatureChecker) *CheckPermissionUseCase {
	return &CheckPermissionUseCase{overrideRepo: overrideRepo, roleChecker: roleChecker}
}

type CheckPermissionInput struct {
	UserID    string
	FeatureID string
	ScopeType string
}

// Execute: el override siempre gana si existe y no ha vencido.
// "Override expirado deja de aplicar automáticamente" -- filtrado en la
// query del repo (FindActiveOverride), no en memoria, para que nunca haya
// ventana donde un override vencido siga otorgando o negando acceso.
func (uc *CheckPermissionUseCase) Execute(ctx context.Context, input CheckPermissionInput) (bool, error) {
	override, err := uc.overrideRepo.FindActiveOverride(ctx, input.UserID, input.FeatureID, input.ScopeType)
	if err != nil {
		return false, err
	}
	if override != nil && !override.IsExpired(time.Now()) {
		return override.IsAllowed, nil
	}
	return uc.roleChecker.HasFeatureViaRole(ctx, input.UserID, input.FeatureID, input.ScopeType)
}