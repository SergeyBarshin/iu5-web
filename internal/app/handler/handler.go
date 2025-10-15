package handler

import (
	"errors"
	"net/http"
	_ "shareholder-app/docs"
	"shareholder-app/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

// RegisterRoutes регистрирует все маршруты API с разделением на группы доступа.
// @title Shareholder App API
// @version 1.0
// @description REST API for Shareholder Management and Dividend Calculation App.
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	// Добавляем Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := router.Group("/api/v1")
	{
		// --- Публичные роуты (доступны всем) ---
		public := api.Group("/")
		{
			// Пользователь
			public.POST("/users/register", h.RegisterUser)
			public.POST("/users/login", h.LoginUser)
			// Акционеры (только чтение)
			public.GET("/shareholders", h.GetShareholders)
			public.GET("/shareholders/:id", h.GetShareholderByID)
		}

		// --- Защищенные роуты (требуют аутентификации) ---
		protected := api.Group("/")
		protected.Use(h.AuthMiddleware(false)) // false = не требует прав модератора
		{
			// Пользователь
			protected.POST("/users/logout", h.LogoutUser)
			protected.GET("/users/me", h.GetMyProfile)
			protected.PUT("/users/me", h.UpdateMyProfile)

			// Акционеры (создание, изменение, добавление в черновик)
			protected.POST("/shareholders/:id/add-to-draft", h.AddShareholderToDraft)

			// Расчеты
			protected.GET("/dividend-calculations/cart", h.GetCartInfo)
			protected.GET("/dividend-calculations", h.GetCalculationsList)
			protected.GET("/dividend-calculations/:id", h.GetCalculationByID)
			protected.PUT("/dividend-calculations/:id", h.UpdateCalculation)
			protected.PUT("/dividend-calculations/:id/submit", h.SubmitCalculation)
			protected.DELETE("/dividend-calculations/:id", h.DeleteCalculation)

			// М-М
			protected.DELETE("/dividend-calculations/:id/shareholders/:shareholder_id", h.DeleteShareholderFromCalculation)
			protected.PUT("/dividend-calculations/:id/shareholders/:shareholder_id", h.UpdateShareholderInCalculation)
		}

		// --- Роуты только для модераторов ---
		moderator := api.Group("/")
		moderator.Use(h.AuthMiddleware(true)) // true = требует прав модератора
		{
			// Акционеры (полный CRUD)
			moderator.POST("/shareholders", h.CreateShareholder)
			moderator.PUT("/shareholders/:id", h.UpdateShareholder)
			moderator.DELETE("/shareholders/:id", h.DeleteShareholder)
			moderator.POST("/shareholders/:id/image", h.UploadShareholderImage)

			// Расчеты (модерация)
			moderator.PUT("/dividend-calculations/:id/moderate", h.ModerateCalculation)
		}
	}

	router.NoRoute(func(ctx *gin.Context) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Page not found"})
	})
}

// errorHandler обрабатывает ошибки и отправляет стандартизированный JSON-ответ
func (h *Handler) errorHandler(ctx *gin.Context, statusCode int, err error) {
	logrus.Error(err.Error())

	// Переопределяем statusCode на основе типа ошибки
	var responseMessage string
	switch {
	case errors.Is(err, repository.ErrNotFound):
		statusCode = http.StatusNotFound
		responseMessage = "Запрашиваемый ресурс не найден"
	case errors.Is(err, repository.ErrAlreadyExists):
		statusCode = http.StatusConflict
		responseMessage = "Ресурс с такими данными уже существует"
	case errors.Is(err, repository.ErrNotAllowed):
		statusCode = http.StatusForbidden
		responseMessage = "Доступ запрещен"
	case errors.Is(err, repository.ErrNoDraft):
		// В зависимости от логики, это может быть не ошибка, а нормальный ответ
		// но для общего обработчика оставим так
		statusCode = http.StatusNotFound
		responseMessage = "Активный черновик не найден"
	default:
		// Если это не одна из наших кастомных ошибок, используем переданный statusCode
		// или устанавливаем 500 по умолчанию
		if statusCode < 400 {
			statusCode = http.StatusInternalServerError
		}
		responseMessage = err.Error() // Для других ошибок показываем их текст
	}

	ctx.JSON(statusCode, gin.H{"error": responseMessage})
}

func (h *Handler) NotFoundPage(ctx *gin.Context)                     {
	ctx.JSON(http.StatusNotFound, gin.H{"error": "Page not found"})
}
