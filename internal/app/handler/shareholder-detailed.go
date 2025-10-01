package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GetShareholderPage возвращает подробную информацию об одном акционере
func (h *Handler) GetShareholderPage(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Errorf("Invalid ID parameter: %v", err)
		ctx.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid ID"})
		return
	}

	shareholder, err := h.Repository.GetShareholderByID(id)
	if err != nil {
		logrus.Errorf("Shareholder not found: %v", err)
		ctx.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Shareholder not found"})
		return
	}

	ctx.HTML(http.StatusOK, "shareholder-detailed.html", gin.H{
		"shareholder": shareholder,
	})
}