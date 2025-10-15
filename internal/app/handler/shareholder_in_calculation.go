// Файл: internal/app/handler/shareholder_in_calculation.go

package handler

import (
	"fmt"
	"net/http"
	"shareholder-app/internal/app/api_types"
	"strconv"

	"github.com/gin-gonic/gin"
)

// DeleteShareholderFromCalculation обрабатывает DELETE /api/dividend-calculations/:id/shareholders/:shareholder_id
// Аналог DeletePlanetFromResearch из референса.
func (h *Handler) DeleteShareholderFromCalculation(ctx *gin.Context) {
	calcIDStr := ctx.Param("id") // Наш роутер использует :id
	calcID, err := strconv.ParseUint(calcIDStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid calculation id parameter"))
		return
	}

	shareholderIDStr := ctx.Param("shareholder_id")
	shareholderID, err := strconv.ParseUint(shareholderIDStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid shareholder id parameter"))
		return
	}

	err = h.Repository.DeleteShareholderFromCalculation(uint(calcID), uint(shareholderID))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// UpdateShareholderInCalculation обрабатывает PUT /api/dividend-calculations/:id/shareholders/:shareholder_id
// Аналог ChangePlanetResearch из референса.
func (h *Handler) UpdateShareholderInCalculation(ctx *gin.Context) {
	calcIDStr := ctx.Param("id")
	calcID, err := strconv.ParseUint(calcIDStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid calculation id parameter"))
		return
	}

	shareholderIDStr := ctx.Param("shareholder_id")
	shareholderID, err := strconv.ParseUint(shareholderIDStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid shareholder id parameter"))
		return
	}

	var req api_types.CalculationItemUpdateRequest
	if err := ctx.BindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	item, err := h.Repository.UpdateShareholderInCalculation(uint(calcID), uint(shareholderID), req)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Конвертируем обновленный item в JSON для ответа
	var finalDividend *float64
	if item.FinalDividend.Valid {
		finalDividend = &item.FinalDividend.Float64
	}
	resp := api_types.CalculationItemResponse{
		Shareholder: api_types.ConvertShareholderToResponse(item.Shareholder),
		Coefficient:   item.Coefficient,
		Fine:          item.Fine,
		FinalDividend: finalDividend,
	}

	ctx.JSON(http.StatusOK, resp)
}