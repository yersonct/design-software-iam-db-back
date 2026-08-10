package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	domaintc "github.com/yersonct/iam-service/internal/domain/trainingcenter"
)

type TrainingCenterHandler struct {
	repo domaintc.Repository
}

func NewTrainingCenterHandler(repo domaintc.Repository) *TrainingCenterHandler {
	return &TrainingCenterHandler{repo: repo}
}

// List maneja GET /training-centers -- catálogo quemado real (SENA).
func (h *TrainingCenterHandler) List(c *gin.Context) {
	c.JSON(http.StatusOK, h.repo.List())
}