package handlers

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/yersonct/iam-service/internal/api/dto"
	applicationscopeoverride "github.com/yersonct/iam-service/internal/application/scopeoverride"
	domainscopeoverride "github.com/yersonct/iam-service/internal/domain/scopeoverride"
)

type ScopeOverrideHandler struct {
	createUC *applicationscopeoverride.CreateOverrideUseCase
	removeUC *applicationscopeoverride.RemoveOverrideUseCase
	listUC   *applicationscopeoverride.ListOverridesUseCase
}

func NewScopeOverrideHandler(
	createUC *applicationscopeoverride.CreateOverrideUseCase,
	removeUC *applicationscopeoverride.RemoveOverrideUseCase,
	listUC *applicationscopeoverride.ListOverridesUseCase,
) *ScopeOverrideHandler {
	return &ScopeOverrideHandler{createUC: createUC, removeUC: removeUC, listUC: listUC}
}

// Create maneja POST /users/{id}/scope-overrides
func (h *ScopeOverrideHandler) Create(c *gin.Context) {
	userID := c.Param("id")
	if !uuidPattern.MatchString(userID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "El ID de usuario debe ser un UUID válido."})
		return
	}

	var req dto.CreateScopeOverrideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload", "message": "Verifica los campos enviados: " + err.Error()})
		return
	}

	grantedByValue, _ := c.Get("user_id")
	grantedBy, _ := grantedByValue.(string)

	o, err := h.createUC.Execute(c.Request.Context(), applicationscopeoverride.CreateOverrideInput{
		UserID:    userID,
		FeatureID: req.FeatureID,
		ScopeType: req.ScopeType,
		IsAllowed: req.IsAllowed,
		Reason:    req.Reason,
		GrantedBy: grantedBy,
		ExpiresAt: req.ExpiresAt,
	})

	if err != nil {
		switch {
		case errors.Is(err, domainscopeoverride.ErrReasonRequired):
			c.JSON(http.StatusBadRequest, gin.H{"error": "reason_required", "message": "Debes indicar un motivo para esta excepción."})
		case errors.Is(err, domainscopeoverride.ErrInvalidScopeType):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_scope_type", "message": "El tipo de alcance (scope_type) no es válido."})
		case errors.Is(err, domainscopeoverride.ErrInvalidExpiresAt):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_expires_at", "message": "La fecha de expiración debe ser futura."})
		case errors.Is(err, domainscopeoverride.ErrUserOrFeatureNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "user_or_feature_not_found", "message": "El usuario o la función indicados no existen."})
		case errors.Is(err, domainscopeoverride.ErrOverrideAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": "scope_override_already_exists", "message": "Ya existe una excepción para este usuario y esta función con este alcance."})
		default:
			log.Printf("[ScopeOverrideHandler.Create] error real: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":          o.ID,
		"user_id":     o.UserID,
		"feature_id":  o.FeatureID,
		"scope_type":  o.ScopeType,
		"is_allowed":  o.IsAllowed,
		"reason":      o.Reason,
		"granted_by":  o.GrantedBy,
		"created_at":  o.CreatedAt,
		"expires_at":  o.ExpiresAt,
	})
}

// Remove maneja DELETE /users/{id}/scope-overrides/{overrideId}
func (h *ScopeOverrideHandler) Remove(c *gin.Context) {
	overrideID := c.Param("overrideId")
	if !uuidPattern.MatchString(overrideID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "El ID debe ser un UUID válido."})
		return
	}

	if err := h.removeUC.Execute(c.Request.Context(), overrideID); err != nil {
		switch {
		case errors.Is(err, domainscopeoverride.ErrOverrideNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "scope_override_not_found", "message": "Esta excepción no existe."})
		default:
			log.Printf("[ScopeOverrideHandler.Remove] error real: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		}
		return
	}
	c.Status(http.StatusNoContent)
}

// List maneja GET /users/{id}/scope-overrides
func (h *ScopeOverrideHandler) List(c *gin.Context) {
	userID := c.Param("id")
	if !uuidPattern.MatchString(userID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "El ID de usuario debe ser un UUID válido."})
		return
	}

	items, err := h.listUC.Execute(c.Request.Context(), userID)
	if err != nil {
		log.Printf("[ScopeOverrideHandler.List] error real: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	now := time.Now()
	response := make([]dto.ScopeOverrideResponse, 0, len(items))
	for _, it := range items {
		response = append(response, dto.ScopeOverrideResponse{
			ID:                 it.ID,
			FeatureID:          it.FeatureID,
			FeatureName:        it.FeatureName,
			FeatureDisplayName: it.FeatureDisplayName,
			ScopeType:          it.ScopeType,
			IsAllowed:          it.IsAllowed,
			Reason:             it.Reason,
			GrantedBy:          it.GrantedBy,
			CreatedAt:          it.CreatedAt,
			ExpiresAt:          it.ExpiresAt,
			IsExpired:          it.ExpiresAt != nil && it.ExpiresAt.Before(now),
		})
	}
	c.JSON(http.StatusOK, response)
}