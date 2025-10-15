// Файл: internal/app/handler/dividend_calculation.go

package handler

import (
	"fmt"
	"net/http"
	"shareholder-app/internal/app/api_types"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// GetCartInfo godoc
// @Summary Get cart info
// @Description Get current draft ID and item count for the authenticated user.
// @Tags calculations
// @Produce json
// @Success 200 {object} api_types.CartInfoResponse
// @Failure 401 {object} map[string]string "Unauthorized"
// @Security BearerAuth
// @Router /dividend-calculations/cart [get]
func (h *Handler) GetCartInfo(ctx *gin.Context) {
	userID, err := GetUserID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}
	draftID, count, err := h.Repository.GetCartInfo(userID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, api_types.CartInfoResponse{
		DraftID: draftID,
		Count:   count,
	})
}

// GetCalculationsList godoc
// @Summary Get list of calculations
// @Description Get all calculations (except draft/deleted). Users see their own, moderators see all.
// @Tags calculations
// @Produce json
// @Param status query string false "Filter by status"
// @Param from_date query string false "Start date for filtering (YYYY-MM-DD)"
// @Param to_date query string false "End date for filtering (YYYY-MM-DD)"
// @Success 200 {array} api_types.CalculationResponse
// @Failure 401 {object} map[string]string "Unauthorized"
// @Security BearerAuth
// @Router /dividend-calculations [get]
func (h *Handler) GetCalculationsList(ctx *gin.Context) {
	userID, err := GetUserID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}
	isModerator, _ := GetUserRole(ctx)

	status := ctx.Query("status")
	dateFromStr := ctx.Query("from_date")
	dateToStr := ctx.Query("to_date")
	var dateFrom, dateTo time.Time
	if dateFromStr != "" {
		dateFrom, err = time.Parse("2006-01-02", dateFromStr)
		if err != nil {
			h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid from_date format"))
			return
		}
	}
	if dateToStr != "" {
		dateTo, err = time.Parse("2006-01-02", dateToStr)
		if err != nil {
			h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid to_date format"))
			return
		}
	}

	calculations, err := h.Repository.GetCalculationsList(dateFrom, dateTo, status, userID, isModerator)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	resp := api_types.ConvertCalculationsToResponse(calculations)
	ctx.JSON(http.StatusOK, resp)
}

// GetCalculationByID godoc
// @Summary Get a calculation by ID
// @Description Get details of a single calculation by its ID.
// @Tags calculations
// @Produce json
// @Param id path int true "Calculation ID"
// @Success 200 {object} api_types.CalculationDetailedResponse
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Security BearerAuth
// @Router /dividend-calculations/{id} [get]
func (h *Handler) GetCalculationByID(ctx *gin.Context) {
	userID, err := GetUserID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}
	isModerator, _ := GetUserRole(ctx)
	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)

	calculation, items, err := h.Repository.GetCalculationWithShareholders(uint(id), userID, isModerator)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	resp := api_types.ConvertCalculationToDetailedResponse(calculation, items)
	ctx.JSON(http.StatusOK, resp)
}

// UpdateCalculation godoc
// @Summary Update a draft calculation
// @Description Update fields of a draft calculation (e.g., total_profit).
// @Tags calculations
// @Accept json
// @Produce json
// @Param id path int true "Calculation ID"
// @Param request body api_types.CalculationUpdateRequest true "Fields to update"
// @Success 200 {object} api_types.CalculationResponse
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden (not creator or not a draft)"
// @Security BearerAuth
// @Router /dividend-calculations/{id} [put]
func (h *Handler) UpdateCalculation(ctx *gin.Context) {
	userID, err := GetUserID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}
	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	var req api_types.CalculationUpdateRequest
	if err := ctx.BindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	calculation, err := h.Repository.UpdateCalculation(uint(id), userID, req)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	resp := api_types.ConvertCalculationToResponse(calculation)
	ctx.JSON(http.StatusOK, resp)
}

// SubmitCalculation godoc
// @Summary Submit a draft calculation
// @Description Change the status of a draft calculation to 'submitted'.
// @Tags calculations
// @Produce json
// @Param id path int true "Calculation ID"
// @Success 200 {object} api_types.CalculationResponse
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden (not creator or not a draft)"
// @Security BearerAuth
// @Router /dividend-calculations/{id}/submit [put]
func (h *Handler) SubmitCalculation(ctx *gin.Context) {
	userID, err := GetUserID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}
	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	calculation, err := h.Repository.SubmitCalculation(uint(id), userID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	resp := api_types.ConvertCalculationToResponse(calculation)
	ctx.JSON(http.StatusOK, resp)
}

// ModerateCalculation godoc
// @Summary Moderate a calculation (Moderator only)
// @Description Complete or reject a 'submitted' calculation. Requires moderator rights.
// @Tags calculations
// @Accept json
// @Produce json
// @Param id path int true "Calculation ID"
// @Param request body api_types.ModerationRequest true "Moderation action"
// @Success 200 {object} api_types.CalculationDetailedResponse
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden (not a moderator or not a submitted calculation)"
// @Security BearerAuth
// @Router /dividend-calculations/{id}/moderate [put]
func (h *Handler) ModerateCalculation(ctx *gin.Context) {
	userID, err := GetUserID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}
	isModerator, _ := GetUserRole(ctx)
	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	var req api_types.ModerationRequest
	if err := ctx.BindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	calculation, err := h.Repository.ModerateCalculation(uint(id), userID, isModerator, req)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	_, items, _ := h.Repository.GetCalculationWithShareholders(uint(id), userID, isModerator)
	resp := api_types.ConvertCalculationToDetailedResponse(calculation, items)
	ctx.JSON(http.StatusOK, resp)
}

// DeleteCalculation godoc
// @Summary Delete a draft calculation
// @Description Logically delete a draft calculation.
// @Tags calculations
// @Produce json
// @Param id path int true "Calculation ID"
// @Success 204 "No Content"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden (not creator or not a draft)"
// @Security BearerAuth
// @Router /dividend-calculations/{id} [delete]
func (h *Handler) DeleteCalculation(ctx *gin.Context) {
	userID, err := GetUserID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}
	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	err = h.Repository.DeleteCalculation(uint(id), userID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}