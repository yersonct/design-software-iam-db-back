package handlers

import (
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"fmt"


	"github.com/gin-gonic/gin"

	"github.com/yersonct/iam-service/internal/api/dto"
	applicationuser "github.com/yersonct/iam-service/internal/application/user"
	domainuser "github.com/yersonct/iam-service/internal/domain/user"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type UserHandler struct {
	createUserUC     *applicationuser.CreateUserUseCase
	listUsersUC      *applicationuser.ListUsersUseCase
	getUserUC        *applicationuser.GetUserUseCase
	updateUserUC     *applicationuser.UpdateUserUseCase
	setUserStatusUC  *applicationuser.SetUserStatusUseCase
	unlockUserUC     *applicationuser.UnlockUserUseCase
}

func NewUserHandler(
	createUserUC *applicationuser.CreateUserUseCase,
	listUsersUC *applicationuser.ListUsersUseCase,
	getUserUC *applicationuser.GetUserUseCase,
	updateUserUC *applicationuser.UpdateUserUseCase,
	setUserStatusUC *applicationuser.SetUserStatusUseCase,
	unlockUserUC *applicationuser.UnlockUserUseCase,
) *UserHandler {
	return &UserHandler{
		createUserUC:    createUserUC,
		listUsersUC:     listUsersUC,
		getUserUC:       getUserUC,
		updateUserUC:    updateUserUC,
		setUserStatusUC: setUserStatusUC,
		unlockUserUC:    unlockUserUC,
	}
}

// Create maneja POST /users.
func (h *UserHandler) Create(c *gin.Context) {
	var req dto.CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_payload",
			"message": err.Error(),
		})
		return
	}

	// Normaliza actor_id: "" o solo espacios cuenta como "no se mandó".
	if req.ActorID != nil {
		trimmed := strings.TrimSpace(*req.ActorID)
		if trimmed == "" {
			req.ActorID = nil
		} else if !uuidPattern.MatchString(trimmed) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_actor_id",
				"message": "El ID del actor debe ser un UUID válido.",
			})
			return
		} else {
			req.ActorID = &trimmed
		}
	}

	out, err := h.createUserUC.Execute(c.Request.Context(), applicationuser.CreateUserInput{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		ActorType: req.ActorType,
		ActorID:   req.ActorID,
	})

	if err != nil {
		h.handleCreateError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.CreateUserResponse{
		ID:        out.ID,
		Email:     out.Email,
		FirstName: out.FirstName,
		LastName:  out.LastName,
		ActorType: out.ActorType,
		ActorID:   out.ActorID,
		IsActive:  out.IsActive,
		EmailSent: out.EmailSent,
	})
}

func (h *UserHandler) handleCreateError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domainuser.ErrEmailAlreadyExists):
		// Criterio de aceptación: email duplicado -> 409 explícito.
		c.JSON(http.StatusConflict, gin.H{
			"error":   "email_already_exists",
			"message": "Ya existe un usuario registrado con ese correo.",
		})

	case errors.Is(err, domainuser.ErrInvalidActorType):
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_actor_type",
			"message": "El tipo de actor debe ser USER, INSTRUCTOR o LEARNER.",
		})

	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Ocurrió un error al crear el usuario.",
		})
	}
}
// List maneja GET /users?page=1&page_size=20&actor_type=INSTRUCTOR&is_active=false
func (h *UserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	in := applicationuser.ListUsersInput{
		Page:     page,
		PageSize: pageSize,
	}

	if actorTypeParam := c.Query("actor_type"); actorTypeParam != "" {
		at := domainuser.ActorType(actorTypeParam)
		if !at.IsValid() {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_actor_type",
				"message": "El tipo de actor debe ser USER, INSTRUCTOR o LEARNER.",
			})
			return
		}
		in.ActorType = &at
	}

	if isActiveParam := c.Query("is_active"); isActiveParam != "" {
		isActive, err := strconv.ParseBool(isActiveParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_is_active",
				"message": "is_active debe ser true o false.",
			})
			return
		}
		in.IsActive = &isActive
	}

	out, err := h.listUsersUC.Execute(c.Request.Context(), in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal_error",
		})
		return
	}

	data := make([]dto.UserSummaryResponse, 0, len(out.Users))
	for _, u := range out.Users {
		data = append(data, toUserSummary(u))
	}

	c.JSON(http.StatusOK, dto.ListUsersResponse{
		Data:       data,
		Page:       out.Page,
		PageSize:   out.PageSize,
		Total:      out.Total,
		TotalPages: out.TotalPages,
	})
}

// GetByID maneja GET /users/{id}
func (h *UserHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	if !uuidPattern.MatchString(id) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_id",
			"message": "El ID debe ser un UUID válido.",
		})
		return
	}

	h.respondUserDetail(c, id)
}

// Me maneja GET /users/me. El id sale del claim "sub" del JWT
// (seteado por middleware.RequireAuth), nunca de la URL.
func (h *UserHandler) Me(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "missing_token",
		})
		return
	}

	userID, _ := userIDValue.(string)

	h.respondUserDetail(c, userID)
}

func (h *UserHandler) respondUserDetail(c *gin.Context, userID string) {
	out, err := h.getUserUC.Execute(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, domainuser.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "user_not_found",
				"message": "El usuario no existe.",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal_error",
		})
		return
	}

	roles := make([]dto.RoleResponse, 0, len(out.Roles))
	for _, r := range out.Roles {
		roles = append(roles, dto.RoleResponse{
			ID:          r.ID,
			Name:        r.Name,
			DisplayName: r.DisplayName,
		})
	}

	c.JSON(http.StatusOK, dto.UserDetailResponse{
		UserSummaryResponse: toUserSummary(out.User),
		Roles:                roles,
	})
}

func toUserSummary(u *domainuser.User) dto.UserSummaryResponse {
	return dto.UserSummaryResponse{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		ActorType: string(u.ActorType),
		ActorID:   u.ActorID,
		IsActive:  u.IsActive,
		IsLocked:  u.IsLocked(time.Now()),
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
	}
}


// Update maneja PUT /users/{id}
func (h *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")

	if !uuidPattern.MatchString(id) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_id",
			"message": "El ID debe ser un UUID válido.",
		})
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_payload",
			"message": err.Error(),
		})
		return
	}

	err := h.updateUserUC.Execute(c.Request.Context(), applicationuser.UpdateUserInput{
		ID:        id,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})

	if err != nil {
		switch {
		case errors.Is(err, domainuser.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "user_not_found",
				"message": "El usuario no existe.",
			})
		case errors.Is(err, domainuser.ErrEmailAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{
				"error":   "email_already_exists",
				"message": "Ya existe un usuario registrado con ese correo.",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "internal_error",
			})
		}
		return
	}

	h.respondUserDetail(c, id)
}

// SetStatus maneja PATCH /users/{id}/status
func (h *UserHandler) SetStatus(c *gin.Context) {
	id := c.Param("id")

	if !uuidPattern.MatchString(id) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_id",
			"message": "El ID debe ser un UUID válido.",
		})
		return
	}

	var req dto.SetUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_payload",
			"message": err.Error(),
		})
		return
	}

	requestingUserID, _ := c.Get("user_id")

	err := h.setUserStatusUC.Execute(c.Request.Context(), applicationuser.SetUserStatusInput{
		TargetUserID:     id,
		RequestingUserID: fmt.Sprint(requestingUserID),
		IsActive:         req.IsActive,
	})

	if err != nil {
		if errors.Is(err, domainuser.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "user_not_found",
				"message": "El usuario no existe.",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal_error",
		})
		return
	}

	h.respondUserDetail(c, id)
}

// Unlock maneja PATCH /users/{id}/unlock
func (h *UserHandler) Unlock(c *gin.Context) {
	id := c.Param("id")

	if !uuidPattern.MatchString(id) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_id",
			"message": "El ID debe ser un UUID válido.",
		})
		return
	}

	if err := h.unlockUserUC.Execute(c.Request.Context(), id); err != nil {
		if errors.Is(err, domainuser.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "user_not_found",
				"message": "El usuario no existe.",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal_error",
		})
		return
	}

	h.respondUserDetail(c, id)
}