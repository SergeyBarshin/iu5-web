package main

import (
	"fmt"

	"shareholder-app/internal/app/config"
	"shareholder-app/internal/app/dsn"
	"shareholder-app/internal/app/handler"
	"shareholder-app/internal/app/repository"
	app "shareholder-app/internal/pkg"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// main инициализирует конфигурацию, репозиторий, хендлеры и запускает приложение
func main() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	
	if err := godotenv.Load(); err != nil {
		logrus.Info("Warning: .env file not found, continuing with environment variables")
	}

	router := gin.Default() // создание нового роутера Gin

	// Загрузка конфигурации приложения
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	// Получение строки подключения к PostgreSQL
	postgresString := dsn.FromEnv()
	if postgresString == "" {
		logrus.Fatal("PostgreSQL DSN string is not configured. Please check your .env file.")
	}
	fmt.Println("DSN string loaded successfully.")

	// Инициализация репозитория
	rep, errRep := repository.NewRepository(postgresString)
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}

	// Создание хендлера с подключённым репозиторием
	hand := handler.NewHandler(rep)

	// Инициализация приложения и запуск сервера
	application := app.NewApp(conf, router, hand)
	application.RunApp()
}