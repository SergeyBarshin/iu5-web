package handler

import (
	"net/http"
	"strconv"
	"strings"

	"shareholder-app/internal/app/ds" // <-- ЗАМЕНИТЕ на имя вашего модуля

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GetShareholdersPage возвращает список акционеров с возможностью поиска
func (h *Handler) GetShareholdersPage(ctx *gin.Context) {
	var shareholders []ds.Shareholder
	var err error
	
	searchQuery := ctx.Query("name")
	if searchQuery == "" {
		shareholders, err = h.Repository.GetShareholders()
	} else {
		shareholders, err = h.Repository.GetShareholdersByName(searchQuery)
	}
	if err != nil {
		logrus.Errorf("Failed to get shareholders: %v", err)
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}

	// Получаем данные для "корзины"
	cartCount := h.Repository.GetCartCount()
	draftID, _ := h.Repository.GetDraftCalculationID(1) // Хардкодим user ID = 1

	ctx.HTML(http.StatusOK, "shareholders-list.html", gin.H{
		"shareholders": shareholders,
		"searchQuery":  searchQuery,
		"cartCount":    cartCount,
		"draftID":      draftID,
	})
}

// AddShareholderToDraft добавляет акционера в расчет-черновик
func (h *Handler) AddShareholderToDraft(ctx *gin.Context) {
	idStr := ctx.PostForm("shareholder_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Errorf("Invalid shareholder ID: %v", err)
		ctx.Redirect(http.StatusFound, "/shareholders?error=invalid_id")
		return
	}

	err = h.Repository.AddShareholderToDraft(id)
	// Игнорируем ошибку дубликата, чтобы пользователь не видел ее
	if err != nil && !strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
		logrus.Errorf("Failed to add shareholder to draft: %v", err)
		ctx.Redirect(http.StatusFound, "/shareholders?error=add_failed")
		return
	}

	// Редирект обратно на главную страницу
	ctx.Redirect(http.StatusFound, "/shareholders")
}