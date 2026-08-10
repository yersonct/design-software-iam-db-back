package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	applicationauthz "github.com/yersonct/iam-service/internal/application/authz"
)

type MeHandler struct {
	getEffectivePermissionsUC *applicationauthz.GetEffectivePermissionsUseCase
}

func NewMeHandler(uc *applicationauthz.GetEffectivePermissionsUseCase) *MeHandler {
	return &MeHandler{getEffectivePermissionsUC: uc}
}

// Me maneja GET /auth/me -- endpoint síncrono que consumen los demás
// microservicios (scheduling-service, academic-management-service, etc.)
// para conocer los permisos EFECTIVOS del usuario dueño del token: roles
// vigentes + features resueltas con overrides ya aplicados. Recalcula
// contra BD en cada llamada -- nunca confía solo en el claim "role" del
// JWT -- así una expiración o un override recién creado se reflejan de
// inmediato, sin esperar a que el token expire o se renueve.
func (h *MeHandler) Me(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing_token", "message": "Se requiere autenticación."})
		return
	}
	userID, _ := userIDValue.(string)

	perms, err := h.getEffectivePermissionsUC.Execute(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	roles := make([]gin.H, 0, len(perms.Roles))
	for _, r := range perms.Roles {
		roles = append(roles, gin.H{
			"role_id": r.RoleID, "role_name": r.RoleName, "role_display_name": r.RoleDisplayName,
			"training_center_id": r.TrainingCenterID, "expires_at": r.ExpiresAt,
		})
	}

	features := make([]gin.H, 0, len(perms.Features))
	for _, f := range perms.Features {
		features = append(features, gin.H{
			"feature_id": f.FeatureID, "feature_code": f.FeatureCode, "feature_name": f.FeatureName,
			"module_code": f.ModuleCode, "module_name": f.ModuleName,
			"scope_type": f.ScopeType, "source": f.Source,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": perms.UserID, "email": perms.Email, "first_name": perms.FirstName, "last_name": perms.LastName,
		"actor_type": perms.ActorType, "is_active": perms.IsActive,
		"roles": roles, "features": features,
	})
}