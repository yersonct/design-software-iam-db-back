package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yersonct/iam-service/internal/api/dto"
	applicationrolefeature "github.com/yersonct/iam-service/internal/application/rolefeature"
	domainrolefeature "github.com/yersonct/iam-service/internal/domain/rolefeature"
)

type RoleFeatureHandler struct {
	assignFeatureUC    *applicationrolefeature.AssignFeatureUseCase
	removeFeatureUC    *applicationrolefeature.RemoveFeatureUseCase
	listRoleFeaturesUC *applicationrolefeature.ListRoleFeaturesUseCase
	batchAssignUC      *applicationrolefeature.BatchAssignFeaturesUseCase // NUEVO
}

func NewRoleFeatureHandler(
	assignFeatureUC *applicationrolefeature.AssignFeatureUseCase,
	removeFeatureUC *applicationrolefeature.RemoveFeatureUseCase,
	listRoleFeaturesUC *applicationrolefeature.ListRoleFeaturesUseCase,
	batchAssignUC *applicationrolefeature.BatchAssignFeaturesUseCase, // NUEVO
) *RoleFeatureHandler {
	return &RoleFeatureHandler{
		assignFeatureUC:    assignFeatureUC,
		removeFeatureUC:    removeFeatureUC,
		listRoleFeaturesUC: listRoleFeaturesUC,
		batchAssignUC:      batchAssignUC,
	}
}

// Assign maneja POST /roles/{id}/features
func (h *RoleFeatureHandler) Assign(c *gin.Context) {
	roleID := c.Param("id")
	if !uuidPattern.MatchString(roleID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "El ID de rol debe ser un UUID válido."})
		return
	}

	var req dto.AssignFeatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Acá cae también un scope_type fuera del enum, gracias al
		// binding:"oneof" -- primera parte del criterio de aceptación.
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload", "message": err.Error()})
		return
	}

	rf, err := h.assignFeatureUC.Execute(c.Request.Context(), applicationrolefeature.AssignFeatureInput{
		RoleID:    roleID,
		FeatureID: req.FeatureID,
		ScopeType: req.ScopeType,
	})

	if err != nil {
		switch {
		case errors.Is(err, domainrolefeature.ErrInvalidScopeType):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_scope_type", "message": "scope_type fuera del conjunto permitido."})
		case errors.Is(err, domainrolefeature.ErrRoleOrFeatureNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "role_or_feature_not_found", "message": "El rol o la feature indicados no existen."})
		case errors.Is(err, domainrolefeature.ErrRoleFeatureAlreadyExists):
			// Segundo criterio de aceptación: no duplicar rol+feature.
			// El frontend ya debe deshabilitar la opción, esto es el
			// respaldo real de la regla (uq_role_feature_role_id_feature_id).
			c.JSON(http.StatusConflict, gin.H{
				"error":   "role_feature_already_exists",
				"message": "Esta feature ya está asignada a este rol.",
			})
		default:
			log.Printf("[RoleFeatureHandler.Assign] error real: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         rf.ID,
		"role_id":    rf.RoleID,
		"feature_id": rf.FeatureID,
		"scope_type": string(rf.ScopeType),
	})
}

// Remove maneja DELETE /roles/{id}/features/{featureId}
func (h *RoleFeatureHandler) Remove(c *gin.Context) {
	roleID := c.Param("id")
	featureID := c.Param("featureId")

	if !uuidPattern.MatchString(roleID) || !uuidPattern.MatchString(featureID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "El ID debe ser un UUID válido."})
		return
	}

	err := h.removeFeatureUC.Execute(c.Request.Context(), roleID, featureID)
	if err != nil {
		switch {
		case errors.Is(err, domainrolefeature.ErrRoleFeatureNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "role_feature_not_found", "message": "Esta feature no está asignada a este rol."})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// List maneja GET /roles/{id}/features
func (h *RoleFeatureHandler) List(c *gin.Context) {
	roleID := c.Param("id")
	if !uuidPattern.MatchString(roleID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "El ID de rol debe ser un UUID válido."})
		return
	}

	items, err := h.listRoleFeaturesUC.Execute(c.Request.Context(), roleID)
	if err != nil {
		log.Printf("[RoleFeatureHandler.List] error real: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	response := make([]dto.RoleFeatureResponse, 0, len(items))
	for _, it := range items {
		response = append(response, dto.RoleFeatureResponse{
			ID:          it.ID,
			FeatureID:   it.FeatureID,
			FeatureCode: it.FeatureCode,
			FeatureName: it.FeatureName,
			ModuleID:    it.ModuleID,
			ModuleCode:  it.ModuleCode,
			ModuleName:  it.ModuleName,
			ScopeType:   string(it.ScopeType),
		})
	}

	c.JSON(http.StatusOK, response)
}
func (h *RoleFeatureHandler) BatchAssign(c *gin.Context) {
	roleID := c.Param("id")
	if !uuidPattern.MatchString(roleID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "El ID de rol debe ser un UUID válido."})
		return
	}

	var req dto.BatchAssignFeaturesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload", "message": err.Error()})
		return
	}

	input := applicationrolefeature.BatchAssignFeaturesInput{RoleID: roleID}
	for _, f := range req.Features {
		input.Features = append(input.Features, applicationrolefeature.BatchAssignFeatureItem{
			FeatureID: f.FeatureID,
			ScopeType: f.ScopeType,
		})
	}

	if err := h.batchAssignUC.Execute(c.Request.Context(), input); err != nil {
		switch {
		case errors.Is(err, domainrolefeature.ErrInvalidScopeType):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_scope_type", "message": "scope_type fuera del conjunto permitido."})
		case errors.Is(err, domainrolefeature.ErrRoleOrFeatureNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "role_or_feature_not_found", "message": "El rol o alguna feature indicada no existen."})
		default:
			log.Printf("[RoleFeatureHandler.BatchAssign] error real: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}