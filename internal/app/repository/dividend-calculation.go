package repository

import (
	"errors"
	"shareholder-app/internal/app/ds"

	"gorm.io/gorm"
)

// GetDraftCalculationID возвращает ID расчета-черновика для пользователя
func (r *Repository) GetDraftCalculationID(userID uint) (int, error) {
	var calculation ds.DividendCalculation
	err := r.db.Where("creator_id = ? AND status = ?", userID, "draft").First(&calculation).Error
	if err != nil {
		return 0, err
	}
	return calculation.ID, nil
}

// GetCalculationDetails возвращает все данные для страницы расчета по его ID
func (r *Repository) GetCalculationDetails(calculationID int) (*ds.DividendCalculation, []ds.ShareholderInCalculation, error) {
	var calculation ds.DividendCalculation
	err := r.db.Where("id = ?", calculationID).First(&calculation).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, nil // Расчета нет
	}
	if err != nil {
		return nil, nil, err
	}

	// Не показываем удаленные расчеты
	if calculation.Status == "deleted" {
		return nil, nil, nil
	}

	// Загружаем связанные элементы (акционеров и их параметры)
	var items []ds.ShareholderInCalculation
	err = r.db.Preload("Shareholder").Where("dividend_calculation_id = ?", calculationID).Find(&items).Error
	if err != nil {
		return nil, nil, err
	}

	return &calculation, items, nil
}

// LogicallyDeleteDraftCalculation выполняет логическое удаление черновика через чистый SQL UPDATE
func (r *Repository) LogicallyDeleteDraftCalculation(userID uint) error {
	query := `UPDATE dividend_calculations SET status = 'deleted' WHERE creator_id = ? AND status = 'draft'`

	result := r.db.Exec(query, userID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no draft calculation found to delete")
	}
	return nil
}
/*
// UpdateCalculationItems обновляет данные (коэффициенты и штрафы) для элементов расчета
func (r *Repository) UpdateCalculationItems(items []ds.ShareholderInCalculation) error {
	// Транзакция, чтобы все обновления были атомарны
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			err := tx.Model(&ds.ShareholderInCalculation{}).
				Where("dividend_calculation_id = ? AND shareholder_id = ?", item.DividendCalculationID, item.ShareholderID).
				Updates(map[string]interface{}{
					"coefficient": item.Coefficient,
					"fine":        item.Fine,
				}).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}*/