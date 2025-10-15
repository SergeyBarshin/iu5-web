// Файл: internal/app/handler/user.go

package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"shareholder-app/internal/app/api_types"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// RegisterUser godoc
// @Summary Register a new user
// @Description Creates a new user account.
// @Tags users
// @Accept json
// @Produce json
// @Param user body api_types.UserRegisterRequest true "User Registration Info"
// @Success 201 {object} api_types.UserResponse "Successfully created user"
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/users/register [post]
func (h *Handler) RegisterUser(ctx *gin.Context) {
	var req api_types.UserRegisterRequest
	if err := ctx.BindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
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

// LoginUser godoc
// @Summary Log in a user
// @Description Authenticates a user and returns a JWT token.
// @Tags users
// @Accept json
// @Produce json
// @Param credentials body api_types.UserLoginRequest true "User Credentials"
// @Success 200 {object} map[string]string "JWT token"
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 401 {object} map[string]string "Invalid credentials"
// @Router /api/v1/users/login [post]
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
	token, err := h.Repository.SignIn(req)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"token": token})
}

// LogoutUser godoc
// @Summary Log out a user
// @Description Invalidates the current user's JWT token by adding it to a blacklist.
// @Tags users
// @Produce json
// @Success 200 {object} map[string]string "Logout status"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Security BearerAuth
// @Router /api/v1/users/logout [post]
func (h *Handler) LogoutUser(ctx *gin.Context) {
	log.Println("[DEBUG] LogoutUser: Handler started.")

	tokenString := extractTokenFromHeader(ctx.Request)
	if tokenString == "" {
		log.Println("[DEBUG] LogoutUser: Token not found in header.")
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("no token provided"))
		return
	}
	log.Printf("[DEBUG] LogoutUser: Extracted token: %s\n", tokenString)

	token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_KEY")), nil
	})

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Println("[DEBUG] LogoutUser: Failed to parse token claims.")
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid token claims"))
		return
	}

	ttl, err := tokenTTLFromClaims(claims)
	if err != nil {
		log.Printf("[DEBUG] LogoutUser: Error getting TTL from claims: %v. Token might be expired.\n", err)
		// Если токен уже истек, это не ошибка, просто выходим.
		ctx.JSON(http.StatusOK, gin.H{"status": "logged_out"})
		return
	}
	log.Printf("[DEBUG] LogoutUser: Calculated TTL: %v\n", ttl)

	err = h.Repository.AddTokenToBlacklist(context.Background(), tokenString, ttl)
	if err != nil {
		log.Printf("[DEBUG] LogoutUser: Error from AddTokenToBlacklist: %v\n", err)
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	log.Println("[DEBUG] LogoutUser: Token successfully added to blacklist.")
	ctx.JSON(http.StatusOK, gin.H{"status": "logged_out"})
}

// GetMyProfile godoc
// @Summary Get current user's profile
// @Description Retrieves the profile information for the authenticated user.
// @Tags users
// @Produce json
// @Success 200 {object} api_types.UserResponse "User profile"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Security BearerAuth
// @Router /api/v1/users/me [get]
func (h *Handler) GetMyProfile(ctx *gin.Context) {
	userID, err := GetUserID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}
	user, err := h.Repository.GetUserByID(userID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, api_types.ConvertUserToResponse(user))
}

// UpdateMyProfile godoc
// @Summary Update current user's profile
// @Description Updates the profile information for the authenticated user.
// @Tags users
// @Accept json
// @Produce json
// @Param user body api_types.UserUpdateRequest true "Fields to update"
// @Success 200 {object} api_types.UserResponse "Updated user profile"
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Security BearerAuth
// @Router /api/v1/users/me [put]
func (h *Handler) UpdateMyProfile(ctx *gin.Context) {
	userID, err := GetUserID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
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