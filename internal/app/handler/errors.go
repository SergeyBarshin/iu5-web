package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// NotFoundPage обрабатывает запросы на несуществующие маршруты.
func (h *Handler) NotFoundPage(ctx *gin.Context) {
	ctx.HTML(http.StatusNotFound, "404.html", nil)
}