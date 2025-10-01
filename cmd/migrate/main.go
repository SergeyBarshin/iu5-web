package main

import (
	"log"

	"shareholder-app/internal/app/ds"
	"shareholder-app/internal/app/dsn"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, continuing with environment variables")
	}

	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}

	log.Println("Starting database migration...")
	err = db.AutoMigrate(
		&ds.User{},
		&ds.Shareholder{},
		&ds.DividendCalculation{},
		&ds.ShareholderInCalculation{},
	)
	if err != nil {
		panic("failed to migrate database")
	}
	log.Println("Database migration completed successfully!")
}