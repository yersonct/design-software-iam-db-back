package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yersonct/iam-service/internal/api/dto"
	applicationcatalog "github.com/yersonct/iam-service/internal/application/catalog"
	domaincatalog "github.com/yersonct/iam-service/internal/domain/catalog"
)

type FeatureHandler struct {
	createFeatureUC          *applicationcatalog.CreateFeatureUseCase
	updateFeatureUC          *applicationcatalog.UpdateFeatureUseCase
	listFeaturesUC           *applicationcatalog.ListFeaturesUseCase
	listFeaturesByModuleUC   *applicationcatalog.ListFeaturesByModuleUseCase
}

func NewFeatureHandler(
	createFeatureUC *applicationcatalog.CreateFeatureUseCase,
	updateFeatureUC *applicationcatalog.UpdateFeatureUseCase,
	listFeaturesUC *applicationcatalog.ListFeaturesUseCase,
	listFeaturesByModuleUC *applicationcatalog.ListFeaturesByModuleUseCase,
) *FeatureHandler {
	return &FeatureHandler{
		createFeatureUC:        createFeatureUC,
		updateFeatureUC:        updateFeatureUC,
		listFeaturesUC:         listFeaturesUC,
		listFeaturesByModuleUC: listFeaturesByModuleUC,
	}
}

// Create maneja POST /features
func (h *FeatureHandler) Create(c *gin.Context) {
	var req dto.CreateFeatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Aquí cae también un action_level fuera del enum, gracias al `oneof`
		// del DTO — es el criterio de aceptación de la historia cumplido.
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload", "message": err.Error()})
		return
	}

	f, err := h.createFeatureUC.Execute(c.Request.Context(), applicationcatalog.CreateFeatureInput{
		ModuleID:    req.ModuleID,
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		ActionLevel: domaincatalog.ActionLevel(req.ActionLevel),
	})

	if err != nil {
		switch {
		case errors.Is(err, domaincatalog.ErrModuleNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "module_not_found", "message": "El módulo indicado no existe."})
		case errors.Is(err, domaincatalog.ErrFeatureCodeAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": "feature_code_already_exists", "message": "Ya existe una feature con ese código."})
		case errors.Is(err, domaincatalog.ErrInvalidActionLevel):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_action_level", "message": "action_level fuera del conjunto permitido."})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		}
		return
	}

	c.JSON(http.StatusCreated, toFeatureResponse(f))
}

// List maneja GET /features
func (h *FeatureHandler) List(c *gin.Context) {
	features, err := h.listFeaturesUC.Execute(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	response := make([]dto.FeatureResponse, 0, len(features))
	for _, f := range features {
		response = append(response, *toFeatureResponse(&f))
	}

	c.JSON(http.StatusOK, response)
}

// Update maneja PUT /features/{id}
func (h *FeatureHandler) Update(c *gin.Context) {
	id := c.Param("id")

	if !uuidPattern.MatchString(id) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "El ID debe ser un UUID válido."})
		return
	}

	var req dto.UpdateFeatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload", "message": err.Error()})
		return
	}

	err := h.updateFeatureUC.Execute(c.Request.Context(), applicationcatalog.UpdateFeatureInput{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		ActionLevel: domaincatalog.ActionLevel(req.ActionLevel),
		IsActive:    *req.IsActive,
	})

	if err != nil {
		switch {
		case errors.Is(err, domaincatalog.ErrFeatureNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "feature_not_found", "message": "La feature no existe."})
		case errors.Is(err, domaincatalog.ErrInvalidActionLevel):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_action_level", "message": "action_level fuera del conjunto permitido."})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// ListByModule maneja GET /modules/{id}/features
func (h *FeatureHandler) ListByModule(c *gin.Context) {
	moduleID := c.Param("id")

	if !uuidPattern.MatchString(moduleID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "El ID debe ser un UUID válido."})
		return
	}

	features, err := h.listFeaturesByModuleUC.Execute(c.Request.Context(), moduleID)
	if err != nil {
		if errors.Is(err, domaincatalog.ErrModuleNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "module_not_found", "message": "El módulo no existe."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	response := make([]dto.FeatureResponse, 0, len(features))
	for _, f := range features {
		response = append(response, *toFeatureResponse(&f))
	}

	c.JSON(http.StatusOK, response)
}

func toFeatureResponse(f *domaincatalog.Feature) *dto.FeatureResponse {
	return &dto.FeatureResponse{
		ID:          f.ID,
		ModuleID:    f.ModuleID,
		Code:        f.Code,
		Name:        f.Name,
		Description: f.Description,
		ActionLevel: string(f.ActionLevel),
		IsActive:    f.IsActive,
		CreatedAt:   f.CreatedAt,
		UpdatedAt:   f.UpdatedAt,
	}
}