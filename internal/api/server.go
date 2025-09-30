package api

import (
	"log"

	"shareholder-app/internal/app/handler"
	"shareholder-app/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func StartServer() {
	log.Println("Запуск сервера")

	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Fatalf("Ошибка инициализации репозитория: %v", err)
	}

	handler := handler.NewHandler(repo)

	r := gin.Default()

	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	r.GET("/shareholders", handler.GetShareholdersPage)
	r.GET("/shareholder/:id", handler.GetShareholderPage)
	r.GET("/dividend-calculation/:id", handler.GetDividendCalculationPage)
	
	log.Println("Сервер успешно запущен на http://localhost:8080")
	r.Run(":8080")
	
	log.Println("Сервер остановлен")
}