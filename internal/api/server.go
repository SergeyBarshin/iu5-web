package api

import (
	"log"

	// Убедитесь, что 'shareholder-app' - это имя вашего модуля из go.mod
	"shareholder-app/internal/app/handler"
	"shareholder-app/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func StartServer() {
	log.Println("Запуск сервера")

	// Инициализируем репозиторий
	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Fatalf("Ошибка инициализации репозитория: %v", err)
	}

	handler := handler.NewHandler(repo)

	// Создаем Gin-роутер
	r := gin.Default()

	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")


	// --- Маршруты (Endpoints) ---

	// Главная страница со списком акционеров
	r.GET("/shareholders", handler.GetShareholdersPage)

	// Страница с детальной информацией об одном акционере
	r.GET("/shareholder/:id", handler.GetShareholderPage)

	// Страница "заявки" для расчета дивидендов
	r.GET("/request", handler.GetRequestPage)


	// Запускаем сервер на порту 8080
	log.Println("Сервер успешно запущен на http://localhost:8080")
	r.Run(":8080")
	
	log.Println("Сервер остановлен")
}