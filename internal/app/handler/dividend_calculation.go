package handler

import (
	"fmt"
	"net/http"
	"shareholder-app/internal/app/api_types"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// GetCartInfo обрабатывает GET /api/dividend-calculations/cart
// Получает "иконку корзины" - ID и количество элементов в черновике.
// Аналог GetResearchCart из референса.
func (h *Handler) GetCartInfo(ctx *gin.Context) {
	draftID, count, err := h.Repository.GetCartInfo()
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Если черновика нет, draftID будет 0.
	// Это нормальная ситуация, а не ошибка.
	ctx.JSON(http.StatusOK, api_types.CartInfoResponse{
		DraftID: draftID,
		Count:   count,
	})
}

// GetCalculationsList обрабатывает GET /api/dividend-calculations
// Получает список расчетов с фильтрацией.
// Аналог GetResearches из референса.
func (h *Handler) GetCalculationsList(ctx *gin.Context) {
	// Парсим query-параметры для фильтрации
	status := ctx.Query("status")
	dateFromStr := ctx.Query("from_date")
	dateToStr := ctx.Query("to_date")

	var dateFrom, dateTo time.Time
	var err error
	if dateFromStr != "" {
		dateFrom, err = time.Parse("2006-01-02", dateFromStr)
		if err != nil {
			h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid from_date format, use YYYY-MM-DD"))
			return
		}
	}
	if dateToStr != "" {
		dateTo, err = time.Parse("2006-01-02", dateToStr)
		if err != nil {
			h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid to_date format, use YYYY-MM-DD"))
			return
		}
	}

	calculations, err := h.Repository.GetCalculationsList(dateFrom, dateTo, status)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Конвертируем модели БД в JSON-ответы
	resp := api_types.ConvertCalculationsToResponse(calculations)
	ctx.JSON(http.StatusOK, resp)
}

// GetCalculationByID обрабатывает GET /api/dividend-calculations/:id
// Получает один расчет со всеми его позициями.
// Аналог GetRsearch из референса.
func (h *Handler) GetCalculationByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid id parameter"))
		return
	}

	calculation, items, err := h.Repository.GetCalculationWithShareholders(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	resp := api_types.ConvertCalculationToDetailedResponse(calculation, items)
	ctx.JSON(http.StatusOK, resp)
}

// UpdateCalculation обрабатывает PUT /api/dividend-calculations/:id
// Обновляет поля черновика (например, общую прибыль).
// Аналог ChangeResearch из референса.
func (h *Handler) UpdateCalculation(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid id parameter"))
		return
	}

	var req api_types.CalculationUpdateRequest
	if err := ctx.BindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	calculation, err := h.Repository.UpdateCalculation(uint(id), req)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Для простоты возвращаем краткий ответ, т.к. список items не менялся
	resp := api_types.ConvertCalculationToResponse(calculation)
	ctx.JSON(http.StatusOK, resp)
}

// SubmitCalculation обрабатывает PUT /api/dividend-calculations/:id/submit
// "Формирует" черновик, меняя его статус на 'submitted'.
// Аналог FormResearch из референса.
func (h *Handler) SubmitCalculation(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid id parameter"))
		return
	}

	calculation, err := h.Repository.SubmitCalculation(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	resp := api_types.ConvertCalculationToResponse(calculation)
	ctx.JSON(http.StatusOK, resp)
}

// ModerateCalculation обрабатывает PUT /api/dividend-calculations/:id/moderate
// "Завершает" или "Отклоняет" расчет, меняя статус.
// Аналог ModerateResearch из референса.
func (h *Handler) ModerateCalculation(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid id parameter"))
		return
	}

	var req api_types.ModerationRequest
	if err := ctx.BindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	calculation, err := h.Repository.ModerateCalculation(uint(id), req)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// После модерации лучше вернуть детальный ответ,
	// так как в случае 'completed' могли рассчитаться дивиденды.
	_, items, _ := h.Repository.GetCalculationWithShareholders(uint(id))
	resp := api_types.ConvertCalculationToDetailedResponse(calculation, items)
	ctx.JSON(http.StatusOK, resp)
}

// DeleteCalculation обрабатывает DELETE /api/dividend-calculations/:id
// Логически удаляет расчет (меняет статус на 'deleted').
// Аналог DeleteResearch из референса.
func (h *Handler) DeleteCalculation(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid id parameter"))
		return
	}

	err = h.Repository.DeleteCalculation(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}