package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yersonct/iam-service/internal/api/dto"
	applicationcatalog "github.com/yersonct/iam-service/internal/application/catalog"
	domaincatalog "github.com/yersonct/iam-service/internal/domain/catalog"
)

type ModuleHandler struct {
	createModuleUC *applicationcatalog.CreateModuleUseCase
	updateModuleUC *applicationcatalog.UpdateModuleUseCase
	listModulesUC  *applicationcatalog.ListModulesUseCase
}

func NewModuleHandler(
	createModuleUC *applicationcatalog.CreateModuleUseCase,
	updateModuleUC *applicationcatalog.UpdateModuleUseCase,
	listModulesUC *applicationcatalog.ListModulesUseCase,
) *ModuleHandler {
	return &ModuleHandler{
		createModuleUC: createModuleUC,
		updateModuleUC: updateModuleUC,
		listModulesUC:  listModulesUC,
	}
}

// Create maneja POST /modules
func (h *ModuleHandler) Create(c *gin.Context) {
	var req dto.CreateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload", "message": err.Error()})
		return
	}

	m, err := h.createModuleUC.Execute(c.Request.Context(), applicationcatalog.CreateModuleInput{
		Code:         req.Code,
		Name:         req.Name,
		Description:  req.Description,
		DisplayOrder: req.DisplayOrder,
		IconKey:      req.IconKey,
	})

	if err != nil {
		if errors.Is(err, domaincatalog.ErrModuleCodeAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "module_code_already_exists",
				"message": "Ya existe un módulo con ese código.",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	c.JSON(http.StatusCreated, toModuleResponse(m))
}

// List maneja GET /modules
func (h *ModuleHandler) List(c *gin.Context) {
	modules, err := h.listModulesUC.Execute(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	response := make([]dto.ModuleResponse, 0, len(modules))
	for _, m := range modules {
		response = append(response, *toModuleResponse(&m))
	}

	c.JSON(http.StatusOK, response)
}

// Update maneja PUT /modules/{id}
func (h *ModuleHandler) Update(c *gin.Context) {
	id := c.Param("id")

	if !uuidPattern.MatchString(id) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "El ID debe ser un UUID válido."})
		return
	}

	var req dto.UpdateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload", "message": err.Error()})
		return
	}

	err := h.updateModuleUC.Execute(c.Request.Context(), applicationcatalog.UpdateModuleInput{
		ID:           id,
		Name:         req.Name,
		Description:  req.Description,
		DisplayOrder: req.DisplayOrder,
		IconKey:      req.IconKey,
		IsActive:     *req.IsActive,
	})

	if err != nil {
		if errors.Is(err, domaincatalog.ErrModuleNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "module_not_found", "message": "El módulo no existe."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	c.Status(http.StatusNoContent)
}

func toModuleResponse(m *domaincatalog.Module) *dto.ModuleResponse {
	return &dto.ModuleResponse{
		ID:           m.ID,
		Code:         m.Code,
		Name:         m.Name,
		Description:  m.Description,
		DisplayOrder: m.DisplayOrder,
		IconKey:      m.IconKey,
		IsActive:     m.IsActive,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}