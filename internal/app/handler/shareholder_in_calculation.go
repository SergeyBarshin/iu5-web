// Файл: internal/app/handler/shareholder_in_calculation.go

package handler

import (
	"net/http"
	"shareholder-app/internal/app/api_types"
	"strconv"

	"github.com/gin-gonic/gin"
)

// DeleteShareholderFromCalculation godoc
// @Summary Delete a shareholder from a draft calculation
// @Description Removes a shareholder from the current user's draft calculation.
// @Tags M-M
// @Produce json
// @Param id path int true "Calculation ID"
// @Param shareholder_id path int true "Shareholder ID"
// @Success 204 "No Content"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden (not creator or not a draft)"
// @Security BearerAuth
// @Router /api/v1/dividend-calculations/{id}/shareholders/{shareholder_id} [delete]
func (h *Handler) DeleteShareholderFromCalculation(ctx *gin.Context) {
	userID, err := GetUserID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}
	calcID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	shareholderID, _ := strconv.ParseUint(ctx.Param("shareholder_id"), 10, 32)

	err = h.Repository.DeleteShareholderFromCalculation(uint(calcID), uint(shareholderID), userID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

// UpdateShareholderInCalculation godoc
// @Summary Update a shareholder's parameters in a draft calculation
// @Description Changes coefficient and fine for a shareholder within the current user's draft.
// @Tags M-M
// @Accept json
// @Produce json
// @Param id path int true "Calculation ID"
// @Param shareholder_id path int true "Shareholder ID"
// @Param request body api_types.CalculationItemUpdateRequest true "Parameters to update"
// @Success 200 {object} api_types.CalculationItemResponse
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden (not creator or not a draft)"
// @Security BearerAuth
// @Router /api/v1/dividend-calculations/{id}/shareholders/{shareholder_id} [put]
func (h *Handler) UpdateShareholderInCalculation(ctx *gin.Context) {
	userID, err := GetUserID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}
	calcID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	shareholderID, _ := strconv.ParseUint(ctx.Param("shareholder_id"), 10, 32)

	var req api_types.CalculationItemUpdateRequest
	if err := ctx.BindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	item, err := h.Repository.UpdateShareholderInCalculation(uint(calcID), uint(shareholderID), userID, req)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	var finalDividend *float64
	if item.FinalDividend.Valid {
		finalDividend = &item.FinalDividend.Float64
	}
	resp := api_types.CalculationItemResponse{
		Shareholder:   api_types.ConvertShareholderToResponse(item.Shareholder),
		Coefficient:   item.Coefficient,
		Fine:          item.Fine,
		FinalDividend: finalDividend,
	}
	ctx.JSON(http.StatusOK, resp)
}