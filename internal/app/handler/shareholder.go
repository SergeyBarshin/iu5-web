// Файл: internal/app/handler/shareholder.go

package handler

import (
	"fmt"
	"net/http"
	"shareholder-app/internal/app/api_types"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GetShareholders обрабатывает GET /api/shareholders
// Получает список акционеров, опционально с фильтром по имени.
func (h *Handler) GetShareholders(ctx *gin.Context) {
	// Получаем query-параметр для фильтрации по имени
	nameFilter := ctx.Query("name")

	shareholders, err := h.Repository.GetShareholdersWithFilter(nameFilter)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Конвертируем модели БД в JSON-ответ
	resp := api_types.ConvertShareholdersToResponse(shareholders)
	ctx.JSON(http.StatusOK, resp)
}

// GetShareholderByID обрабатывает GET /api/shareholders/:id
// Получает одного акционера по его ID.
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

// CreateShareholder обрабатывает POST /api/shareholders
// Создает нового акционера.
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

	// Устанавливаем заголовок Location, как в референсе, для REST-совместимости
	ctx.Header("Location", fmt.Sprintf("/api/shareholders/%d", shareholder.ID))
	ctx.JSON(http.StatusCreated, api_types.ConvertShareholderToResponse(shareholder))
}

// UpdateShareholder обрабатывает PUT /api/shareholders/:id
// Обновляет существующего акционера.
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

// DeleteShareholder обрабатывает DELETE /api/shareholders/:id
// Удаляет акционера.
func (h *Handler) DeleteShareholder(ctx *gin.Context) {
	logrus.Info("DeleteShareholder handler: started") // <-- Лог 1: Начало работы

	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid id parameter: %w", err))
		return
	}

	logrus.Infof("DeleteShareholder handler: parsed shareholder ID: %d", id) // <-- Лог 2: ID получен

	logrus.Info("DeleteShareholder handler: calling repository method...") // <-- Лог 3: Перед вызовом репозитория

	err = h.Repository.DeleteShareholder(uint(id))

	logrus.Info("DeleteShareholder handler: repository method finished.") // <-- Лог 4: После вызова репозитория

	if err != nil {
		logrus.Errorf("DeleteShareholder handler: repository returned an error: %v", err)
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	logrus.Info("DeleteShareholder handler: sending 204 No Content response") // <-- Лог 5: Успешное завершение

	// При успешном удалении возвращаем статус 204 No Content
	ctx.Status(http.StatusNoContent)
}

// UploadShareholderImage обрабатывает POST /api/shareholders/:id/image
// Загружает изображение для акционера.
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

	ctx.JSON(http.StatusOK, gin.H{
		"image_url": shareholder.ImageURL.String,
	})
}

// AddShareholderToDraft обрабатывает POST /api/shareholders/:id/add-to-draft
// Добавляет акционера в черновик расчета.
func (h *Handler) AddShareholderToDraft(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid id parameter: %w", err))
		return
	}

	err = h.Repository.AddShareholderToDraftCalculation(uint(id))
	if err != nil {
		// Здесь мы передаем ошибку как есть, т.к. наш errorHandler уже умеет
		// обрабатывать ErrNotFound, ErrAlreadyExists и т.д.
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// В референсе возвращался созданный research, но по заданию нам достаточно
	// просто вернуть подтверждение успеха.
	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "shareholder added to draft"})
}