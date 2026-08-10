package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	applicationaudit "github.com/yersonct/iam-service/internal/application/audit"
	domainaudit "github.com/yersonct/iam-service/internal/domain/audit"
)

type AuditHandler struct {
	listLoginUC *applicationaudit.ListLoginAuditUseCase
}

func NewAuditHandler(uc *applicationaudit.ListLoginAuditUseCase) *AuditHandler {
	return &AuditHandler{listLoginUC: uc}
}

func (h *AuditHandler) ListLogins(c *gin.Context) {
	filter := domainaudit.ListFilter{}

	if email := c.Query("email"); email != "" {
		filter.Email = &email
	}
	if successStr := c.Query("success"); successStr != "" {
		success := successStr == "true"
		filter.Success = &success
	}
	if fromStr := c.Query("from"); fromStr != "" {
		if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
			filter.From = &t
		}
	}
	if toStr := c.Query("to"); toStr != "" {
		if t, err := time.Parse(time.RFC3339, toStr); err == nil {
			filter.To = &t
		}
	}
	filter.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	filter.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.listLoginUC.Execute(c.Request.Context(), filter)
	if err != nil {
		log.Printf("error en ListLogins: %v", err) // TEMPORAL para diagnóstico
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	items := make([]gin.H, 0, len(result.Items))
	for _, it := range result.Items {
		items = append(items, gin.H{
			"id": it.ID, "email_attempted": it.EmailAttempted,
			"outcome": it.Outcome, "attempted_at": it.AttemptedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items, "total": result.Total,
		"page": result.Page, "page_size": result.PageSize,
	})
}