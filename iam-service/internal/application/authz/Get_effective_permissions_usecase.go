package authz

import (
	"context"
	"time"

	domainrolefeature "github.com/yersonct/iam-service/internal/domain/rolefeature"
	domainscopeoverride "github.com/yersonct/iam-service/internal/domain/scopeoverride"
	domainuser "github.com/yersonct/iam-service/internal/domain/user"
	domainuserrole "github.com/yersonct/iam-service/internal/domain/userrole"
)

type EffectiveRole struct {
	RoleID           string
	RoleName         string
	RoleDisplayName  string
	TrainingCenterID *string
	ExpiresAt        *time.Time
}

// EffectiveFeature: feature YA resuelta (rol + overrides aplicados).
// Source indica de dónde salió el permiso final, útil para debugging
// desde los otros microservicios que consumen /auth/me.
type EffectiveFeature struct {
	FeatureID   string
	FeatureCode string
	FeatureName string
	ModuleCode  string
	ModuleName  string
	ScopeType   string
	Source      string // "role" | "override"
}

type EffectivePermissions struct {
	UserID    string
	Email     string
	FirstName string
	LastName  string
	ActorType string
	IsActive  bool
	Roles     []EffectiveRole
	Features  []EffectiveFeature
}

type GetEffectivePermissionsUseCase struct {
	userRepo     domainuser.Repository
	userRoleRepo domainuserrole.Repository
	roleFeature  domainrolefeature.Repository
	overrideRepo domainscopeoverride.Repository
	clock        func() time.Time
}

func NewGetEffectivePermissionsUseCase(
	userRepo domainuser.Repository,
	userRoleRepo domainuserrole.Repository,
	roleFeature domainrolefeature.Repository,
	overrideRepo domainscopeoverride.Repository,
) *GetEffectivePermissionsUseCase {
	return &GetEffectivePermissionsUseCase{
		userRepo: userRepo, userRoleRepo: userRoleRepo,
		roleFeature: roleFeature, overrideRepo: overrideRepo,
		clock: time.Now,
	}
}

func (uc *GetEffectivePermissionsUseCase) Execute(ctx context.Context, userID string) (*EffectivePermissions, error) {
	now := uc.clock()

	u, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	allUserRoles, err := uc.userRoleRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Filtro de vigencia: un rol con expires_at vencido deja de contar aquí
	// mismo, en cada resolución -- no solo se marca visualmente en algún
	// lado. Este es el corazón del criterio de expiración de la historia.
	var activeRoles []domainuserrole.UserRoleItem
	roles := make([]EffectiveRole, 0, len(allUserRoles))
	for _, ur := range allUserRoles {
		if ur.ExpiresAt != nil && ur.ExpiresAt.Before(now) {
			continue
		}
		activeRoles = append(activeRoles, ur)
		roles = append(roles, EffectiveRole{
			RoleID: ur.RoleID, RoleName: ur.RoleName, RoleDisplayName: ur.RoleDisplayName,
			TrainingCenterID: ur.TrainingCenterID, ExpiresAt: ur.ExpiresAt,
		})
	}

	// Base: features otorgadas por los roles vigentes.
	featureMap := make(map[string]EffectiveFeature)
	for _, ur := range activeRoles {
		items, err := uc.roleFeature.ListByRole(ctx, ur.RoleID)
		if err != nil {
			return nil, err
		}
		for _, it := range items {
			featureMap[it.FeatureID] = EffectiveFeature{
				FeatureID: it.FeatureID, FeatureCode: it.FeatureCode, FeatureName: it.FeatureName,
				ModuleCode: it.ModuleCode, ModuleName: it.ModuleName,
				ScopeType: string(it.ScopeType), Source: "role",
			}
		}
	}

	// Overrides vigentes SIEMPRE ganan sobre el rol -- mismo principio que
	// CheckPermissionUseCase, aplicado aquí a la lista completa: allow
	// agrega/reemplaza la feature, deny la quita aunque el rol la otorgara.
	overrides, err := uc.overrideRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, ov := range overrides {
		if ov.ExpiresAt != nil && ov.ExpiresAt.Before(now) {
			continue // override vencido: se ignora, como si no existiera
		}
		if ov.IsAllowed {
			// Nota: en ScopeOverrideItem, "FeatureName" es en realidad el
			// código de la feature y "FeatureDisplayName" el nombre legible
			// -- mismo patrón invertido que ya tenías en ese struct.
			featureMap[ov.FeatureID] = EffectiveFeature{
				FeatureID: ov.FeatureID, FeatureCode: ov.FeatureName, FeatureName: ov.FeatureDisplayName,
				ScopeType: ov.ScopeType, Source: "override",
			}
		} else {
			delete(featureMap, ov.FeatureID)
		}
	}

	features := make([]EffectiveFeature, 0, len(featureMap))
	for _, f := range featureMap {
		features = append(features, f)
	}

	return &EffectivePermissions{
		UserID: u.ID, Email: u.Email, FirstName: u.FirstName, LastName: u.LastName,
		ActorType: string(u.ActorType), IsActive: u.IsActive,
		Roles: roles, Features: features,
	}, nil
}