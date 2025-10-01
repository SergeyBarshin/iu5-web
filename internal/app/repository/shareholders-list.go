package repository

import (
	"database/sql"
	"fmt"
	"shareholder-app/internal/app/ds"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// GetShareholders возвращает всех акционеров
func (r *Repository) GetShareholders() ([]ds.Shareholder, error) {
	var shareholders []ds.Shareholder
	err := r.db.Find(&shareholders).Error
	return shareholders, err
}

// GetShareholdersByName ищет акционеров по имени
func (r *Repository) GetShareholdersByName(name string) ([]ds.Shareholder, error) {
	var shareholders []ds.Shareholder
	err := r.db.Where("name ILIKE ?", "%"+name+"%").Find(&shareholders).Error
	return shareholders, err
}

// GetCartCount возвращает количество акционеров в расчете-черновике для пользователя creator_id=1
func (r *Repository) GetCartCount() int64 {
	var calculationID int
	var count int64
	creatorID := 1 // Хардкодим ID пользователя для примера

	err := r.db.Model(&ds.DividendCalculation{}).
		Where("creator_id = ? AND status = ?", creatorID, "draft").
		Select("id").
		First(&calculationID).Error
	if err != nil {
		return 0 // Если черновика нет, в корзине 0 элементов
	}

	err = r.db.Model(&ds.ShareholderInCalculation{}).
		Where("dividend_calculation_id = ?", calculationID).
		Count(&count).Error
	if err != nil {
		logrus.Errorf("Error counting shareholders in calculation: %v", err)
		return 0
	}
	return count
}

// AddShareholderToDraft добавляет акционера в расчет-черновик
func (r *Repository) AddShareholderToDraft(shareholderID int) error {
	var calculation ds.DividendCalculation
	creatorID := 1    // Хардкодим ID пользователя
	moderatorID := 2  // Хардкодим ID модератора для примера
	now := time.Now()

	// Ищем существующий черновик или создаем новый
	err := r.db.Where("creator_id = ? AND status = ?", creatorID, "draft").First(&calculation).Error
	if err != nil && err == gorm.ErrRecordNotFound {
		// Создаем новый расчет-черновик, если он не найден
		calculation = ds.DividendCalculation{
			Status:      "draft",
			CreatedAt:   now,
			CreatorID:   uint(creatorID),
			ModeratorID: sql.NullInt64{Int64: int64(moderatorID), Valid: true},
		}
		if err = r.db.Create(&calculation).Error; err != nil {
			return fmt.Errorf("failed to create draft calculation: %w", err)
		}
	} else if err != nil {
		return err // Другая ошибка при поиске
	}

	// Создаем связь "акционер <-> расчет" с параметрами по умолчанию
	record := ds.ShareholderInCalculation{
		DividendCalculationID: calculation.ID,
		ShareholderID:         shareholderID,
		Coefficient:           1.0, // Значение по умолчанию
		Fine:                  0.0, // Значение по умолчанию
	}
	// Используем FirstOrCreate, чтобы избежать дубликатов
	err = r.db.Where(ds.ShareholderInCalculation{
		DividendCalculationID: record.DividendCalculationID,
		ShareholderID:         record.ShareholderID,
	}).FirstOrCreate(&record).Error
	if err != nil {
		return fmt.Errorf("error adding shareholder to calculation: %w", err)
	}

	return nil
}