package minioClient

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"shareholder-app/internal/app/ds"
	"strconv"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Client - это наша обертка над стандартным клиентом MinIO
type Client struct {
	*minio.Client
	bucketName string
}

// New создает и настраивает новый клиент для MinIO.
func New() (*Client, error) {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	useSSL, _ := strconv.ParseBool(os.Getenv("MINIO_USE_SSL"))
	bucketName := os.Getenv("MINIO_BUCKET_NAME")

	if endpoint == "" || accessKey == "" || secretKey == "" || bucketName == "" {
		return nil, fmt.Errorf("minio environment variables are not fully set")
	}

	// Инициализация клиента MinIO
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	// Проверяем, существует ли бакет, и создаем его, если нет
	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, err
	}
	if !exists {
		err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, err
		}
		// Устанавливаем публичную политику для бакета, чтобы изображения были доступны по URL
		policy := fmt.Sprintf(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::%s/*"]}]}`, bucketName)
		err = minioClient.SetBucketPolicy(ctx, bucketName, policy)
		if err != nil {
			return nil, err
		}
	}

	return &Client{
		Client:     minioClient,
		bucketName: bucketName,
	}, nil
}

// UploadImage загружает файл изображения в MinIO.
// Имя объекта генерируется на основе ID акционера и расширения файла.
func (c *Client) UploadImage(ctx context.Context, file *multipart.FileHeader, shareholder ds.Shareholder) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// Генерируем уникальное имя для объекта
	ext := filepath.Ext(file.Filename)
	objectName := fmt.Sprintf("shareholder-%d%s", shareholder.ID, ext)

	// Загружаем объект в бакет
	_, err = c.PutObject(ctx, c.bucketName, objectName, src, file.Size, minio.PutObjectOptions{
		ContentType: file.Header.Get("Content-Type"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload object: %w", err)
	}

	// Возвращаем полный URL к загруженному файлу
	url := fmt.Sprintf("http://%s/%s/%s", os.Getenv("MINIO_ENDPOINT"), c.bucketName, objectName)
	return url, nil
}

// DeleteImage удаляет объект из MinIO.
func (c *Client) DeleteImage(ctx context.Context, objectName string) error {
	err := c.RemoveObject(ctx, c.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	return nil
}


// "http://localhost:9000/images/shareholder-1.jpg" -> "shareholder-1.jpg"
func (c *Client) GetObjectNameFromURL(url string) string {
	// Мы можем просто взять последнюю часть URL после слэша
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}