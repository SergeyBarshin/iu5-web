package handler

import (
	"errors"
	"net/http"
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

// RegisterRoutes регистрирует все маршруты API в точности по заданию
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api")
	{
		// --- Домен "Пользователь" (User) ---
		// POST /api/users/register - Регистрация
		// POST /api/users/login - Аутентификация
		// POST /api/users/logout - Деавторизация
		// GET /api/users/me - Получение данных о "себе" (личный кабинет)
		// PUT /api/users/me - Изменение данных о "себе" (личный кабинет)
		users := api.Group("/users")
		{
			users.POST("/register", h.RegisterUser)
			users.POST("/login", h.LoginUser)
			users.POST("/logout", h.LogoutUser) // Деавторизация
			users.GET("/me", h.GetMyProfile)     // Личный кабинет
			users.PUT("/me", h.UpdateMyProfile)  // Личный кабинет
		}

		// --- Домен "Акционеры" (Shareholders) ---
		shareholders := api.Group("/shareholders")
		{
			shareholders.GET("", h.GetShareholders)            // GET список с фильтрацией
			shareholders.POST("", h.CreateShareholder)           // POST добавление (без изображения)
			shareholders.GET("/:id", h.GetShareholderByID)       // GET одна запись
			shareholders.PUT("/:id", h.UpdateShareholder)        // PUT изменение
			shareholders.DELETE("/:id", h.DeleteShareholder)       // DELETE удаление
			shareholders.POST("/:id/image", h.UploadShareholderImage) // POST добавление изображения
			shareholders.POST("/:id/add-to-draft", h.AddShareholderToDraft) // POST добавления в заявку-черновик

		}

		// --- Домен "Расчеты" (Dividend Calculations) ---
		calculations := api.Group("/dividend-calculations")
		{
			calculations.GET("/cart", h.GetCartInfo)                   // GET иконки корзины
			calculations.GET("", h.GetCalculationsList)            // GET список (с фильтрацией)
			calculations.GET("/:id", h.GetCalculationByID)           // GET одна запись
			calculations.PUT("/:id", h.UpdateCalculation)            // PUT изменения полей
			calculations.PUT("/:id/submit", h.SubmitCalculation)           // PUT сформировать
			calculations.PUT("/:id/moderate", h.ModerateCalculation)       // PUT завершить/отклонить
			calculations.DELETE("/:id", h.DeleteCalculation)             // DELETE удаление (логическое)
		}

		// --- Домен "М-М" (Элементы в расчете) ---
		// Используем вложенные роуты для REST-совместимости
		calculationItems := api.Group("/dividend-calculations/:id/shareholders/:shareholder_id")
		{
			calculationItems.DELETE("", h.DeleteShareholderFromCalculation) // DELETE удаление из заявки
			calculationItems.PUT("", h.UpdateShareholderInCalculation)    // PUT изменение в м-м
		}
	}
	
	router.NoRoute(h.NotFoundPage) // Обработчик 404
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
