// Файл: internal/app/handler/user.go

package handler

import (
	"fmt"
	"net/http"
	"shareholder-app/internal/app/api_types"

	"github.com/gin-gonic/gin"
)

// RegisterUser обрабатывает POST /api/users/register
func (h *Handler) RegisterUser(ctx *gin.Context) {
	var req api_types.UserRegisterRequest
	if err := ctx.BindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Валидация
	if req.Login == "" || req.Password == "" {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("login and password are required"))
		return
	}

	user, err := h.Repository.RegisterUser(req)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, api_types.ConvertUserToResponse(user))
}

// LoginUser обрабатывает POST /api/users/login
func (h *Handler) LoginUser(ctx *gin.Context) {
	var req api_types.UserLoginRequest
	if err := ctx.BindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if req.Login == "" || req.Password == "" {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("login and password are required"))
		return
	}

	user, err := h.Repository.LoginUser(req)
	if err != nil {
		// Ошибка "invalid login or password" уже сформирована в репозитории
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}

	ctx.JSON(http.StatusOK, api_types.ConvertUserToResponse(user))
}

// LogoutUser обрабатывает POST /api/users/logout
func (h *Handler) LogoutUser(ctx *gin.Context) {
	h.Repository.SignOut()
	// Возвращаем константного пользователя по умолчанию
	h.Repository.SetUserID(1) 
	ctx.JSON(http.StatusOK, gin.H{"status": "logged_out"})
}

// GetMyProfile обрабатывает GET /api/users/me
func (h *Handler) GetMyProfile(ctx *gin.Context) {
	userID := h.Repository.GetUserID()
	if userID == 0 {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("user not authenticated"))
		return
	}

	user, err := h.Repository.GetUserByID(userID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, api_types.ConvertUserToResponse(user))
}

// UpdateMyProfile обрабатывает PUT /api/users/me
func (h *Handler) UpdateMyProfile(ctx *gin.Context) {
	userID := h.Repository.GetUserID()
	if userID == 0 {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("user not authenticated"))
		return
	}

	var req api_types.UserUpdateRequest
	if err := ctx.BindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	user, err := h.Repository.UpdateProfile(userID, req)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, api_types.ConvertUserToResponse(user))
}