package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"shareholder-app/internal/app/minioClient"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"log"

	"gorm.io/gorm/logger" // <-- Добавьте этот импорт
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrNotAllowed    = errors.New("not allowed")
	ErrNoDraft       = errors.New("draft not found")
)

type Repository struct {
	db *gorm.DB
	mc *minioClient.Client
	rd *redis.Client
}

func NewRepository(dsn string, mc *minioClient.Client) (*Repository, error) {
	newLogger := logger.New(
        log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
        logger.Config{
            SlowThreshold:             time.Second, // Порог медленных запросов
            LogLevel:                  logger.Info, // Уровень логгирования
            IgnoreRecordNotFoundError: true,        // Не логировать ошибки "запись не найдена"
            Colorful:                  true,        // Цветной вывод
        },
    )


	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: newLogger})
	if err != nil {
		return nil, err
	}

	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	rd := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: "",
		DB:       0,
	})

	if _, err := rd.Ping(context.Background()).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Repository{
		db: db,
		mc: mc,
		rd: rd,
	}, nil
}

func blacklistKeyForToken(tokenString string) string {
	h := sha256.Sum256([]byte(tokenString))
	return "blacklist:" + hex.EncodeToString(h[:])
}

func (r *Repository) AddTokenToBlacklist(ctx context.Context, tokenString string, ttl time.Duration) error {
	if ttl <= 0 {
		return nil
	}
	key := blacklistKeyForToken(tokenString)
	return r.rd.Set(ctx, key, "1", ttl).Err()
}

func (r *Repository) IsTokenBlacklisted(ctx context.Context, tokenString string) (bool, error) {
	key := blacklistKeyForToken(tokenString)
	val, err := r.rd.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return val > 0, nil
}