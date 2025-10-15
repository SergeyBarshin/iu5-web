package repository

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"shareholder-app/internal/app/api_types"
	"shareholder-app/internal/app/ds"
	"time"

	"gorm.io/gorm"
)

// GetShareholdersWithFilter возвращает список акционеров.
func (r *Repository) GetShareholdersWithFilter(nameFilter string) ([]ds.Shareholder, error) {
	var shareholders []ds.Shareholder
	query := r.db.Order("name ASC") // Сортируем по имени для предсказуемого результата
	if nameFilter != "" {
		query = query.Where("name ILIKE ?", "%"+nameFilter+"%")
	}
	err := query.Find(&shareholders).Error
	return shareholders, err
}

// GetShareholderByID возвращает одного акционера по ID.
func (r *Repository) GetShareholderByID(id uint) (ds.Shareholder, error) {
	var shareholder ds.Shareholder
	err := r.db.First(&shareholder, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.Shareholder{}, fmt.Errorf("%w: shareholder with id %d", ErrNotFound, id)
		}
		return ds.Shareholder{}, err
	}
	return shareholder, nil
}

// CreateShareholder создает нового акционера.
func (r *Repository) CreateShareholder(req api_types.ShareholderRequest) (ds.Shareholder, error) {
	// Валидация, как в референсе (проверка на отрицательные значения и т.д.)
	if req.Name == "" {
		return ds.Shareholder{}, errors.New("shareholder name cannot be empty")
	}
	if req.Share <= 0 || req.Share > 100 {
		return ds.Shareholder{}, errors.New("share must be between 0 and 100")
	}

	newShareholder := ds.Shareholder{
		Name:        req.Name,
		Description: req.Description,
		Share:       req.Share,
	}
	err := r.db.Create(&newShareholder).Error
	return newShareholder, err
}

// UpdateShareholder обновляет данные акционера.
func (r *Repository) UpdateShareholder(id uint, req api_types.ShareholderRequest) (ds.Shareholder, error) {
	// Сначала находим существующего акционера
	shareholder, err := r.GetShareholderByID(id)
	if err != nil {
		return ds.Shareholder{}, err // Ошибка, если не найден
	}

	// Валидация
	if req.Name == "" {
		return ds.Shareholder{}, errors.New("shareholder name cannot be empty")
	}
	if req.Share <= 0 || req.Share > 100 {
		return ds.Shareholder{}, errors.New("share must be between 0 and 100")
	}

	// Обновляем поля и сохраняем
	shareholder.Name = req.Name
	shareholder.Description = req.Description
	shareholder.Share = req.Share

	err = r.db.Save(&shareholder).Error
	return shareholder, err
}

// DeleteShareholder удаляет акционера. Удаление изображения встроено.
func (r *Repository) DeleteShareholder(id uint) error {
	shareholder, err := r.GetShareholderByID(id)
	if err != nil {
		return err // Акционер не найден
	}

	// Если у акционера есть изображение, удаляем его из MinIO
	if shareholder.ImageURL.Valid && shareholder.ImageURL.String != "" {
		objectName := r.mc.GetObjectNameFromURL(shareholder.ImageURL.String)
		err := r.mc.DeleteImage(context.Background(), objectName)
		if err != nil {
			// Логируем ошибку, но не прерываем операцию, так как удаление из БД важнее.
			fmt.Printf("warning: failed to delete image '%s' from minio: %v\n", objectName, err)
		}
	}
	// Физическое удаление записи из БД
	return r.db.Unscoped().Delete(&ds.Shareholder{}, id).Error
}

// UploadShareholderImage загружает изображение в MinIO и обновляет запись в БД.
func (r *Repository) UploadShareholderImage(id uint, file *multipart.FileHeader) (ds.Shareholder, error) {
	shareholder, err := r.GetShareholderByID(id)
	if err != nil {
		return ds.Shareholder{}, err
	}

	// Перед загрузкой нового изображения удаляем старое, если оно есть.
	// Это соответствует требованию "старое изображение заменяется/удаляется".
	if shareholder.ImageURL.Valid && shareholder.ImageURL.String != "" {
		objectName := r.mc.GetObjectNameFromURL(shareholder.ImageURL.String)
		err := r.mc.DeleteImage(context.Background(), objectName)
		if err != nil {
			fmt.Printf("warning: failed to delete old image '%s' from minio: %v\n", objectName, err)
		}
	}

	// Загружаем новое изображение в MinIO
	imageURL, err := r.mc.UploadImage(context.Background(), file, shareholder)
	if err != nil {
		return ds.Shareholder{}, fmt.Errorf("failed to upload image to minio: %w", err)
	}

	// Обновляем URL в записи БД
	shareholder.ImageURL.String = imageURL
	shareholder.ImageURL.Valid = true
	err = r.db.Save(&shareholder).Error
	if err != nil {
		// Если не удалось сохранить в БД, пытаемся откатить загрузку из MinIO
		objectName := r.mc.GetObjectNameFromURL(imageURL)
		_ = r.mc.DeleteImage(context.Background(), objectName)
		return ds.Shareholder{}, fmt.Errorf("failed to save image URL to database: %w", err)
	}

	return shareholder, nil
}

// AddShareholderToDraftCalculation добавляет акционера в черновик расчета.
func (r *Repository) AddShareholderToDraftCalculation(shareholderID uint) error {
	creatorID := r.GetUserID()
	if creatorID == 0 {
		return fmt.Errorf("%w: user is not authenticated", ErrNotAllowed)
	}

	// Получаем или создаем черновик расчета для текущего пользователя
	draft, err := r.getOrCreateDraftCalculation(creatorID)
	if err != nil {
		return err
	}

	// Убедимся, что акционер с таким ID существует
	if _, err := r.GetShareholderByID(shareholderID); err != nil {
		return err
	}

	// Проверяем, не добавлен ли этот акционер уже в расчет
	var existingLink ds.ShareholderInCalculation
	err = r.db.Where("dividend_calculation_id = ? AND shareholder_id = ?", draft.ID, shareholderID).First(&existingLink).Error
	if err == nil {
		return fmt.Errorf("%w: shareholder is already in the draft calculation", ErrAlreadyExists)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err // Другая ошибка БД
	}

	// Создаем новую запись в M-M таблице с дефолтными значениями
	newLink := ds.ShareholderInCalculation{
		DividendCalculationID: draft.ID,
		ShareholderID:         int(shareholderID),
		Coefficient:           1.0, // Значение по умолчанию
		Fine:                  0.0, // Значение по умолчанию
	}

	return r.db.Create(&newLink).Error
}

func (r *Repository) getOrCreateDraftCalculation(creatorID uint) (ds.DividendCalculation, error) {
	var calculation ds.DividendCalculation
	err := r.db.Where("creator_id = ? AND status = ?", creatorID, "draft").First(&calculation).Error

	if err == nil {
		// Черновик найден, возвращаем его
		return calculation, nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Черновик не найден, создаем новый
		newDraft := ds.DividendCalculation{
			Status:    "draft",
			CreatorID: creatorID,
			CreatedAt: time.Now(),
		}
		if errCreate := r.db.Create(&newDraft).Error; errCreate != nil {
			return ds.DividendCalculation{}, errCreate
		}
		return newDraft, nil
	}

	// Любая другая ошибка
	return ds.DividendCalculation{}, err
}