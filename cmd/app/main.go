// Файл: cmd/app/main.go
package main

import (
	"log"

	"shareholder-app/internal/app/config"
	"shareholder-app/internal/app/dsn"
	"shareholder-app/internal/app/handler"
	"shareholder-app/internal/app/minioClient"
	"shareholder-app/internal/app/repository"
	app "shareholder-app/internal/pkg"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	// код логгера, godotenv, router, config
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	
	if err := godotenv.Load(); err != nil {
		logrus.Info("Warning: .env file not found, continuing with environment variables")
	}

	router := gin.Default()

	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	postgresString := dsn.FromEnv()
	if postgresString == "" {
		logrus.Fatal("PostgreSQL DSN string is not configured. Please check your .env file.")
	}
	log.Println("DSN string loaded successfully.")

	// Инициализация клиента MinIO
	mc, errMc := minioClient.New()
	if errMc != nil {
		logrus.Fatalf("error initializing minio client: %v", errMc)
	}
	log.Println("MinIO client initialized successfully.")

	// Инициализация репозитория (теперь с MinIO)
	rep, errRep := repository.NewRepository(postgresString, mc)
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}
	log.Println("Repository initialized successfully.")
	
	// Создание хендлера (теперь без MinIO)
	hand := handler.NewHandler(rep)

	application := app.NewApp(conf, router, hand)
	application.RunApp()
}