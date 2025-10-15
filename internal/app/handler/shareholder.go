package handler

import (
	"fmt"
	"net/http"
	"shareholder-app/internal/app/api_types"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetShareholders godoc
// @Summary Get list of shareholders
// @Description Get all shareholders with optional name filter. Publicly accessible.
// @Tags shareholders
// @Produce json
// @Param name query string false "Filter by shareholder name (case-insensitive)"
// @Success 200 {array} api_types.ShareholderResponse
// @Router /api/v1/shareholders [get]
func (h *Handler) GetShareholders(ctx *gin.Context) {
	nameFilter := ctx.Query("name")
	shareholders, err := h.Repository.GetShareholdersWithFilter(nameFilter)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	resp := api_types.ConvertShareholdersToResponse(shareholders)
	ctx.JSON(http.StatusOK, resp)
}

// GetShareholderByID godoc
// @Summary Get a shareholder by ID
// @Description Get details of a single shareholder by its ID. Publicly accessible.
// @Tags shareholders
// @Produce json
// @Param id path int true "Shareholder ID"
// @Success 200 {object} api_types.ShareholderResponse
// @Failure 400 {object} map[string]string "Invalid ID format"
// @Failure 404 {object} map[string]string "Shareholder not found"
// @Router /api/v1/shareholders/{id} [get]
func (h *Handler) GetShareholderByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid id parameter: %w", err))
		return
	}
	shareholder, err := h.Repository.GetShareholderByID(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	resp := api_types.ConvertShareholderToResponse(shareholder)
	ctx.JSON(http.StatusOK, resp)
}

// CreateShareholder godoc
// @Summary Create a new shareholder (Moderator only)
// @Description Add a new shareholder to the database. Requires moderator rights.
// @Tags shareholders
// @Accept json
// @Produce json
// @Param shareholder body api_types.ShareholderRequest true "Shareholder object"
// @Success 201 {object} api_types.ShareholderResponse
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden (not a moderator)"
// @Security BearerAuth
// @Router /api/v1/shareholders [post]
func (h *Handler) CreateShareholder(ctx *gin.Context) {
	var req api_types.ShareholderRequest
	if err := ctx.BindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	shareholder, err := h.Repository.CreateShareholder(req)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.Header("Location", fmt.Sprintf("/api/v1/shareholders/%d", shareholder.ID))
	ctx.JSON(http.StatusCreated, api_types.ConvertShareholderToResponse(shareholder))
}

// UpdateShareholder godoc
// @Summary Update a shareholder (Moderator only)
// @Description Update an existing shareholder's data. Requires moderator rights.
// @Tags shareholders
// @Accept json
// @Produce json
// @Param id path int true "Shareholder ID"
// @Param shareholder body api_types.ShareholderRequest true "Shareholder object"
// @Success 200 {object} api_types.ShareholderResponse
// @Failure 400 {object} map[string]string "Invalid ID or request body"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden (not a moderator)"
// @Security BearerAuth
// @Router /api/v1/shareholders/{id} [put]
func (h *Handler) UpdateShareholder(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid id parameter: %w", err))
		return
	}
	var req api_types.ShareholderRequest
	if err := ctx.BindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	shareholder, err := h.Repository.UpdateShareholder(uint(id), req)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, api_types.ConvertShareholderToResponse(shareholder))
}

// DeleteShareholder godoc
// @Summary Delete a shareholder (Moderator only)
// @Description Logically delete a shareholder by its ID. Requires moderator rights.
// @Tags shareholders
// @Produce json
// @Param id path int true "Shareholder ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string "Invalid ID format"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden (not a moderator)"
// @Security BearerAuth
// @Router /api/v1/shareholders/{id} [delete]
func (h *Handler) DeleteShareholder(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid id parameter: %w", err))
		return
	}
	err = h.Repository.DeleteShareholder(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

// UploadShareholderImage godoc
// @Summary Upload an image for a shareholder (Moderator only)
// @Description Upload an image and associate it with a shareholder. Requires moderator rights.
// @Tags shareholders
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "Shareholder ID"
// @Param image formData file true "Image file"
// @Success 200 {object} map[string]string "Image URL"
// @Failure 400 {object} map[string]string "Invalid ID or file"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden (not a moderator)"
// @Security BearerAuth
// @Router /api/v1/shareholders/{id}/image [post]
func (h *Handler) UploadShareholderImage(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid id parameter: %w", err))
		return
	}
	file, err := ctx.FormFile("image")
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("image file is required in 'image' form field: %w", err))
		return
	}
	shareholder, err := h.Repository.UploadShareholderImage(uint(id), file)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"image_url": shareholder.ImageURL.String})
}

// AddShareholderToDraft godoc
// @Summary Add a shareholder to the draft calculation
// @Description Adds a shareholder to the current user's draft calculation.
// @Tags shareholders
// @Produce json
// @Param id path int true "Shareholder ID to add"
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} map[string]string "Invalid ID format"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Security BearerAuth
// @Router /api/v1/shareholders/{id}/add-to-draft [post]
func (h *Handler) AddShareholderToDraft(ctx *gin.Context) {
	// --- ИЗМЕНЕНИЕ ---
	userID, err := GetUserID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}

	shareholderIDStr := ctx.Param("id")
	shareholderID, err := strconv.ParseUint(shareholderIDStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid id parameter: %w", err))
		return
	}

	err = h.Repository.AddShareholderToDraftCalculation(uint(shareholderID), userID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "shareholder added to draft"})
}