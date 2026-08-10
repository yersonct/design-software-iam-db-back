package handlers

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/yersonct/iam-service/internal/api/dto"
	applicationuserrole "github.com/yersonct/iam-service/internal/application/userrole"
	domainuserrole "github.com/yersonct/iam-service/internal/domain/userrole"
)

type UserRoleHandler struct {
	assignRoleUC    *applicationuserrole.AssignRoleUseCase
	removeRoleUC    *applicationuserrole.RemoveRoleUseCase
	listUserRolesUC *applicationuserrole.ListUserRolesUseCase
}

func NewUserRoleHandler(
	assignRoleUC *applicationuserrole.AssignRoleUseCase,
	removeRoleUC *applicationuserrole.RemoveRoleUseCase,
	listUserRolesUC *applicationuserrole.ListUserRolesUseCase,
) *UserRoleHandler {
	return &UserRoleHandler{assignRoleUC: assignRoleUC, removeRoleUC: removeRoleUC, listUserRolesUC: listUserRolesUC}
}

// Assign maneja POST /users/{id}/roles
func (h *UserRoleHandler) Assign(c *gin.Context) {
	userID := c.Param("id")
	if !uuidPattern.MatchString(userID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "El ID de usuario debe ser un UUID válido."})
		return
	}

	var req dto.AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload", "message": err.Error()})
		return
	}

	assignedByValue, _ := c.Get("user_id")
	assignedBy, _ := assignedByValue.(string)

	ur, err := h.assignRoleUC.Execute(c.Request.Context(), applicationuserrole.AssignRoleInput{
		UserID:           userID,
		RoleID:           req.RoleID,
		TrainingCenterID: req.TrainingCenterID,
		AssignedBy:       assignedBy,
		ExpiresAt:        req.ExpiresAt,
	})

	if err != nil {
		switch {
		case errors.Is(err, domainuserrole.ErrInvalidTrainingCenter):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_training_center", "message": "El centro de formación indicado no existe."})
		case errors.Is(err, domainuserrole.ErrInvalidExpiresAt):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_expires_at", "message": "La fecha de expiración debe ser futura."})
		case errors.Is(err, domainuserrole.ErrUserOrRoleNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "user_or_role_not_found", "message": "El usuario o el rol indicados no existen."})
		case errors.Is(err, domainuserrole.ErrUserRoleAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": "user_role_already_exists", "message": "Este usuario ya tiene este rol asignado en este centro de formación."})
		default:
			log.Printf("[UserRoleHandler.Assign] error real: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":                 ur.ID,
		"user_id":            ur.UserID,
		"role_id":            ur.RoleID,
		"training_center_id": ur.TrainingCenterID,
		"assigned_by":        ur.AssignedBy,
		"assigned_at":        ur.AssignedAt,
		"expires_at":         ur.ExpiresAt,
	})
}

// Remove maneja DELETE /users/{id}/roles/{roleId}?training_center_id=...
// El query param es OPCIONAL: si se omite, borra la asignación "global"
// (training_center_id IS NULL); si se manda, borra esa específica.
func (h *UserRoleHandler) Remove(c *gin.Context) {
	userID := c.Param("id")
	roleID := c.Param("roleId")

	if !uuidPattern.MatchString(userID) || !uuidPattern.MatchString(roleID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "El ID debe ser un UUID válido."})
		return
	}

	var trainingCenterID *string
	if tc := c.Query("training_center_id"); tc != "" {
		if !uuidPattern.MatchString(tc) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "training_center_id debe ser un UUID válido."})
			return
		}
		trainingCenterID = &tc
	}

	if err := h.removeRoleUC.Execute(c.Request.Context(), userID, roleID, trainingCenterID); err != nil {
		switch {
		case errors.Is(err, domainuserrole.ErrUserRoleNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "user_role_not_found", "message": "Esta asignación no existe."})
		default:
			log.Printf("[UserRoleHandler.Remove] error real: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		}
		return
	}
	c.Status(http.StatusNoContent)
}

// List maneja GET /users/{id}/roles
func (h *UserRoleHandler) List(c *gin.Context) {
	userID := c.Param("id")
	if !uuidPattern.MatchString(userID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "El ID de usuario debe ser un UUID válido."})
		return
	}

	items, err := h.listUserRolesUC.Execute(c.Request.Context(), userID)
	if err != nil {
		log.Printf("[UserRoleHandler.List] error real: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	now := time.Now()
	response := make([]dto.UserRoleResponse, 0, len(items))
	for _, it := range items {
		response = append(response, dto.UserRoleResponse{
			ID:               it.ID,
			RoleID:           it.RoleID,
			RoleName:         it.RoleName,
			RoleDisplayName:  it.RoleDisplayName,
			TrainingCenterID: it.TrainingCenterID,
			AssignedBy:       it.AssignedBy,
			AssignedAt:       it.AssignedAt,
			ExpiresAt:        it.ExpiresAt,
			IsExpired:        it.ExpiresAt != nil && it.ExpiresAt.Before(now),
		})
	}
	c.JSON(http.StatusOK, response)
}