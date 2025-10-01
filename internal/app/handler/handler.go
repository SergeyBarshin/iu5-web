package handler

import (
	"shareholder-app/internal/app/repository" // <-- ЗАМЕНИТЕ на имя вашего модуля

	"github.com/gin-gonic/gin"
)

// Handler обрабатывает HTTP-запросы и хранит ссылку на репозиторий
type Handler struct {
	Repository *repository.Repository
}

// NewHandler создает новый Handler с подключенным репозиторием
func NewHandler(r *repository.Repository) *Handler {
	return &Handler{Repository: r}
}

// RegisterRoutes регистрирует все маршруты для обработки HTTP-запросов
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	// --- GET-запросы (отображение страниц) ---
	router.GET("/shareholders", h.GetShareholdersPage)
	router.GET("/shareholder/:id", h.GetShareholderPage)
	router.GET("/dividend-calculation/:id", h.GetDividendCalculationPage)

	// --- POST-запросы (действия) ---
	router.POST("/add-shareholder-to-request", h.AddShareholderToDraft)
	router.POST("/delete-request", h.LogicallyDeleteDraft)

	//router.POST("/dividend-calculation/:id/update", h.UpdateCalculation)
}

// RegisterStatic регистрирует статические файлы и шаблоны
func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("./templates/*")
	router.Static("/static", "./resources")
}