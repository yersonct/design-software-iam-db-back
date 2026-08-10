package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	applicationauthz "github.com/yersonct/iam-service/internal/application/authz"
)

type AuthzHandler struct {
	checkPermissionUC *applicationauthz.CheckPermissionUseCase
}

func NewAuthzHandler(checkPermissionUC *applicationauthz.CheckPermissionUseCase) *AuthzHandler {
	return &AuthzHandler{checkPermissionUC: checkPermissionUC}
}

// Check maneja GET /authz/check?user_id=...&feature_id=...&scope_type=...
//
// Endpoint de verificación: responde si ese usuario puede usar esa feature
// con ese scope, en este momento. Prioriza el override sobre el rol (ver
// CheckPermissionUseCase.Execute) -- es la forma de probar en vivo el
// criterio "el motor de autorización debe consultar overrides con
// prioridad sobre lo que da el rol" y "override expirado deja de aplicar
// automáticamente", sin tener que leer el código.
func (h *AuthzHandler) Check(c *gin.Context) {
	userID := c.Query("user_id")
	featureID := c.Query("feature_id")
	scopeType := c.Query("scope_type")

	if !uuidPattern.MatchString(userID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_user_id", "message": "user_id debe ser un UUID válido."})
		return
	}
	if !uuidPattern.MatchString(featureID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_feature_id", "message": "feature_id debe ser un UUID válido."})
		return
	}
	if scopeType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_scope_type", "message": "scope_type es obligatorio."})
		return
	}

	allowed, err := h.checkPermissionUC.Execute(c.Request.Context(), applicationauthz.CheckPermissionInput{
		UserID:    userID,
		FeatureID: featureID,
		ScopeType: scopeType,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":    userID,
		"feature_id": featureID,
		"scope_type": scopeType,
		"allowed":    allowed,
	})
}