package handler

import (
	"net/http"
	"strconv"

	"shareholder-app/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

type DividendResult struct {
	Shareholder repository.Shareholder
	Coefficient float64
	Fine        float64
	Dividend    float64
}

func (h *Handler) GetShareholdersPage(ctx *gin.Context) {
	var shareholders []repository.Shareholder
	var err error
	searchQuery := ctx.Query("query")
	if searchQuery == "" {
		shareholders, err = h.Repository.GetShareholders()
	} else {
		shareholders, err = h.Repository.GetShareholdersByName(searchQuery)
	}
	if err != nil {
		logrus.Errorf("Ошибка: %v", err)
	}

	count, _ := h.Repository.GetTotalRequestItemsCount()

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"shareholders": shareholders,
		"query":        searchQuery,
		"Count":        count,
	})
}

func (h *Handler) GetShareholderPage(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.String(http.StatusBadRequest, "Некорректный ID")
		return
	}
	shareholder, err := h.Repository.GetShareholder(id)
	if err != nil {
		ctx.String(http.StatusNotFound, "Акционер не найден")
		return
	}
	ctx.HTML(http.StatusOK, "shareholder.html", gin.H{
		"shareholder": shareholder,
	})
}
func (h *Handler) GetRequestPage(ctx *gin.Context) {
	// id заявки из URL
	idStr := ctx.Param("id")
	requestID, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.String(http.StatusBadRequest, "Некорректный ID заявки")
		return
	}

	// существует ли такая заявка
	_, err = h.Repository.GetRequest(requestID)
	if err != nil {
		ctx.String(http.StatusNotFound, "Заявка не найдена")
		return
	}

	totalProfitStr := ctx.DefaultQuery("total_profit", "10000")
	totalProfit, _ := strconv.ParseFloat(totalProfitStr, 64)

	requestItems, _ := h.Repository.GetRequestItemsByRequestID(requestID)

	coeffsStr := ctx.QueryArray("coeffs")
	finesStr := ctx.QueryArray("fines")
	idsStr := ctx.QueryArray("ids")

	results := make([]DividendResult, 0, len(requestItems))
	for i, item := range requestItems {
		shareholder, _ := h.Repository.GetShareholder(item.ShareholderID)

		coeff := item.Coefficient
		fine := item.Fine
		
		if len(coeffsStr) > i && len(finesStr) > i && len(idsStr) > i && idsStr[i] == strconv.Itoa(shareholder.ID) {
			coeff, _ = strconv.ParseFloat(coeffsStr[i], 64)
			fine, _ = strconv.ParseFloat(finesStr[i], 64)
		}

		dividend := (totalProfit * (shareholder.Share / 100)) * coeff - fine

		results = append(results, DividendResult{
			Shareholder: shareholder,
			Coefficient: coeff,
			Fine:        fine,
			Dividend:    dividend,
		})
	}
	
	shareholdersCount := len(requestItems)

	ctx.HTML(http.StatusOK, "request.html", gin.H{
		"requestID":         requestID,
		"results":           results,
		"totalProfit":       totalProfit,
		"shareholdersCount": shareholdersCount,
	})
}