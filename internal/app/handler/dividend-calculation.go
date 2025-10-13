package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GetDividendCalculationPage отображает страницу расчета дивидендов
func (h *Handler) GetDividendCalculationPage(ctx *gin.Context) {
	idStr := ctx.Param("id")
	calculationID, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.Redirect(http.StatusFound, "/shareholders")
		return
	}

	calculation, items, err := h.Repository.GetCalculationDetails(calculationID)
	if err != nil {
		logrus.Errorf("Failed to get calculation details: %v", err)
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}
	
	// Если расчет не найден или удален, возвращаем 404
	if calculation == nil {
		ctx.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Calculation not found or has been deleted"})
		return
	}
	
	// Если в "корзине" пусто, редирект на главную
	if len(items) == 0 {
		ctx.Redirect(http.StatusFound, "/shareholders")
		return
	}

	ctx.HTML(http.StatusOK, "dividend-calculation.html", gin.H{
		"calculation": *calculation,
		"items":       items,
	})
}

// LogicallyDeleteDraft выполняет логическое удаление черновика
func (h *Handler) LogicallyDeleteDraft(ctx *gin.Context) {
	userID := uint(1) // Хардкодим ID пользователя
	
	err := h.Repository.LogicallyDeleteDraftCalculation(userID)
	if err != nil {
		logrus.Errorf("Failed to delete draft: %v", err)
		// Можно показать страницу с ошибкой, но редирект проще
	}

	ctx.Redirect(http.StatusFound, "/shareholders")
}

/*
// UpdateCalculation обрабатывает форму со страницы расчета для обновления данных
func (h *Handler) UpdateCalculation(ctx *gin.Context) {
	idStr := ctx.Param("id")
	calculationID, _ := strconv.Atoi(idStr)
	
	// Парсим данные из формы
	formShareholderIDs := ctx.PostFormArray("shareholder_id")
	formCoefficients := ctx.PostFormArray("coefficient")
	formFines := ctx.PostFormArray("fine")
	
	var itemsToUpdate []ds.ShareholderInCalculation
	for i := range formShareholderIDs {
		shareholderID, _ := strconv.Atoi(formShareholderIDs[i])
		coefficient, _ := strconv.ParseFloat(formCoefficients[i], 64)
		fine, _ := strconv.ParseFloat(formFines[i], 64)
		
		itemsToUpdate = append(itemsToUpdate, ds.ShareholderInCalculation{
			DividendCalculationID: calculationID,
			ShareholderID:         shareholderID,
			Coefficient:           coefficient,
			Fine:                  fine,
		})
	}
	
	// Вызываем метод репозитория для обновления
	if err := h.Repository.UpdateCalculationItems(itemsToUpdate); err != nil {
		logrus.Errorf("Failed to update calculation items: %v", err)
		// В случае ошибки просто редиректим обратно с сообщением
		ctx.Redirect(http.StatusFound, fmt.Sprintf("/dividend-calculation/%d?error=update_failed", calculationID))
		return
	}
	
	// Редирект на ту же страницу для отображения обновленных данных
	ctx.Redirect(http.StatusFound, fmt.Sprintf("/dividend-calculation/%d", calculationID))
}*/