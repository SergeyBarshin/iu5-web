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
	shareholderName := ctx.Query("name")
	if shareholderName == "" {
		shareholders, err = h.Repository.GetShareholders()
	} else {
		shareholders, err = h.Repository.GetShareholdersByName(shareholderName)
	}
	if err != nil {
		logrus.Errorf("Ошибка: %v", err)
	}
	count, _ := h.Repository.GetTotalRequestItemsCount()
	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"shareholders": shareholders,
		"name":         shareholderName,
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


func (h *Handler) GetDividendCalculationPage(ctx *gin.Context) {
	idStr := ctx.Param("id")
	requestID, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.String(http.StatusBadRequest, "Некорректный ID заявки")
		return
	}

	companyProfitStr := ctx.DefaultQuery("company_profit", "10000")
	companyProfit, _ := strconv.ParseFloat(companyProfitStr, 64)

	updates := make(map[int]repository.UpdatePayload)
	idsStr := ctx.QueryArray("shareholder_id")
	coeffsStr := ctx.QueryArray("adjustment_coefficient")
	finesStr := ctx.QueryArray("shareholder_fine")

	for i, idStr := range idsStr {
		id, _ := strconv.Atoi(idStr)
		if i < len(coeffsStr) && i < len(finesStr) {
			coeff, _ := strconv.ParseFloat(coeffsStr[i], 64)
			fine, _ := strconv.ParseFloat(finesStr[i], 64)
			updates[id] = repository.UpdatePayload{
				Coefficient: coeff,
				Fine:        fine,
			}
		}
	}

	updatedRequestItems, err := h.Repository.UpdateAndCalculateRequestItems(requestID, companyProfit, updates)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "Ошибка при расчете дивидендов")
		return
	}

	results := make([]DividendResult, 0, len(updatedRequestItems))
	for _, item := range updatedRequestItems {
		shareholder, _ := h.Repository.GetShareholder(item.ShareholderID)
		results = append(results, DividendResult{
			Shareholder: shareholder,
			Coefficient: item.Coefficient,
			Fine:        item.Fine,
			Dividend:    item.CalculatedDividend,
		})
	}
	
	shareholdersCount := len(updatedRequestItems)

	ctx.HTML(http.StatusOK, "request.html", gin.H{
		"requestID":         requestID,
		"results":           results,
		"totalProfit":       companyProfit,
		"shareholdersCount": shareholdersCount,
	})
}