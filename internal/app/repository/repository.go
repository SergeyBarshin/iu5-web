package repository

import (
	"errors"
	"shareholder-app/internal/app/minioClient"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrNotAllowed    = errors.New("not allowed")
	ErrNoDraft       = errors.New("draft not found")
)

type Repository struct {
	db     *gorm.DB
	mc     *minioClient.Client // Клиент MinIO как в референсе
	userId uint                // ID "авторизованного" пользователя
}

// NewRepository принимает и minioClient
func NewRepository(dsn string, mc *minioClient.Client) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	// Устанавливаем пользователя по умолчанию, как требует задание
	// Например, пользователь с ID=1 будет нашим создателем
	return &Repository{
		db:     db,
		mc:     mc,
		userId: 1, // Пользователь-создатель по умолчанию
	}, nil
}

// --- Методы для управления "сессией" ---

func (r *Repository) GetUserID() uint {
	return r.userId
}

func (r *Repository) SetUserID(id uint) {
	r.userId = id
}

func (r *Repository) SignOut() {
	r.userId = 0 // "Разлогиниваемся", сбрасывая ID
}