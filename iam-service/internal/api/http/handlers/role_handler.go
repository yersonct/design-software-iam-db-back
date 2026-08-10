package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yersonct/iam-service/internal/api/dto"
	applicationrole "github.com/yersonct/iam-service/internal/application/role"
	domainrole "github.com/yersonct/iam-service/internal/domain/role"
)

type RoleHandler struct {
	createRoleUC *applicationrole.CreateRoleUseCase
	updateRoleUC *applicationrole.UpdateRoleUseCase
	deleteRoleUC *applicationrole.DeleteRoleUseCase
	listRolesUC  *applicationrole.ListRolesUseCase
}

func NewRoleHandler(
	createRoleUC *applicationrole.CreateRoleUseCase,
	updateRoleUC *applicationrole.UpdateRoleUseCase,
	deleteRoleUC *applicationrole.DeleteRoleUseCase,
	listRolesUC *applicationrole.ListRolesUseCase,
) *RoleHandler {
	return &RoleHandler{
		createRoleUC: createRoleUC,
		updateRoleUC: updateRoleUC,
		deleteRoleUC: deleteRoleUC,
		listRolesUC:  listRolesUC,
	}
}

// Create maneja POST /roles
func (h *RoleHandler) Create(c *gin.Context) {
	var req dto.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload", "message": err.Error()})
		return
	}

	r, err := h.createRoleUC.Execute(c.Request.Context(), applicationrole.CreateRoleInput{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: derefString(req.Description),
	})

	if err != nil {
		switch {
		case errors.Is(err, domainrole.ErrRoleNameAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": "role_name_already_exists", "message": "Ya existe un rol con ese nombre."})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		}
		return
	}

	c.JSON(http.StatusCreated, toRoleResponse(r))
}

// List maneja GET /roles
func (h *RoleHandler) List(c *gin.Context) {
	roles, err := h.listRolesUC.Execute(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	response := make([]dto.RoleCatalogResponse, 0, len(roles))
	for _, r := range roles {
		response = append(response, *toRoleResponse(&r))
	}

	c.JSON(http.StatusOK, response)
}

// Update maneja PUT /roles/{id}
func (h *RoleHandler) Update(c *gin.Context) {
	id := c.Param("id")

	if !uuidPattern.MatchString(id) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "El ID debe ser un UUID válido."})
		return
	}

	var req dto.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload", "message": err.Error()})
		return
	}

	err := h.updateRoleUC.Execute(c.Request.Context(), applicationrole.UpdateRoleInput{
		ID:          id,
		DisplayName: req.DisplayName,
		Description: derefString(req.Description),
	})

	if err != nil {
		switch {
		case errors.Is(err, domainrole.ErrRoleNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "role_not_found", "message": "El rol no existe."})
		case errors.Is(err, domainrole.ErrSystemRoleNotEditable):
			// Criterio de aceptación: mensaje explícito, no un genérico 500/400.
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "system_role_not_editable",
				"message": "No se puede editar un rol de sistema.",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// Delete maneja DELETE /roles/{id}
func (h *RoleHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if !uuidPattern.MatchString(id) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "El ID debe ser un UUID válido."})
		return
	}

	err := h.deleteRoleUC.Execute(c.Request.Context(), id)

	if err != nil {
		switch {
		case errors.Is(err, domainrole.ErrRoleNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "role_not_found", "message": "El rol no existe."})
		case errors.Is(err, domainrole.ErrSystemRoleNotDeletable):
			// Mismo criterio de aceptación, para el caso de eliminar.
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "system_role_not_deletable",
				"message": "No se puede eliminar un rol de sistema.",
			})
		case errors.Is(err, domainrole.ErrRoleInUse):
			c.JSON(http.StatusConflict, gin.H{
				"error":   "role_in_use",
				"message": "El rol tiene usuarios asignados y no se puede eliminar.",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func toRoleResponse(r *domainrole.Role) *dto.RoleCatalogResponse {
	var description *string
	if r.Description != "" {
		description = &r.Description
	}

	return &dto.RoleCatalogResponse{
		ID:           r.ID,
		Name:         r.Name,
		DisplayName:  r.DisplayName,
		Description:  description,
		IsSystemRole: r.IsSystemRole,
		CreatedAt:    r.CreatedAt,
	}
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}