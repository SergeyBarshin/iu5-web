package repository

import (
	"fmt"
	"shareholder-app/internal/app/api_types"
	"shareholder-app/internal/app/ds"
)

// checkDraftAccess - это вспомогательная функция для проверки, что пользователь
// имеет право редактировать указанный черновик.
func (r *Repository) checkDraftAccess(calculationID uint) error {
	var calculation ds.DividendCalculation
	err := r.db.Where("id = ?", calculationID).First(&calculation).Error
	if err != nil {
		return fmt.Errorf("%w: calculation not found", ErrNotFound)
	}

	if calculation.Status != "draft" {
		return fmt.Errorf("%w: can only modify a draft calculation", ErrNotAllowed)
	}

	if calculation.CreatorID != r.GetUserID() {
		return fmt.Errorf("%w: you are not the creator of this draft", ErrNotAllowed)
	}
	return nil
}

// DeleteShareholderFromCalculation удаляет акционера из черновика расчета.
// Аналог DeletePlanetFromResearch из референса.
func (r *Repository) DeleteShareholderFromCalculation(calculationID, shareholderID uint) error {
	// Сначала проверяем, имеет ли пользователь доступ к этому черновику
	if err := r.checkDraftAccess(calculationID); err != nil {
		return err
	}

	// Выполняем удаление записи из M-M таблицы
	result := r.db.Where("dividend_calculation_id = ? AND shareholder_id = ?", calculationID, shareholderID).
		Delete(&ds.ShareholderInCalculation{})

	if result.Error != nil {
		return result.Error
	}

	// Проверяем, была ли запись действительно удалена
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: shareholder not found in this calculation draft", ErrNotFound)
	}

	return nil
}

// UpdateShareholderInCalculation изменяет данные акционера (коэффициент, штраф) в черновике.
func (r *Repository) UpdateShareholderInCalculation(calculationID, shareholderID uint, req api_types.CalculationItemUpdateRequest) (ds.ShareholderInCalculation, error) {
	// Проверяем доступ к черновику
	if err := r.checkDraftAccess(calculationID); err != nil {
		return ds.ShareholderInCalculation{}, err
	}

	// Валидация входных данных
	if req.Coefficient <= 0 {
		return ds.ShareholderInCalculation{}, fmt.Errorf("coefficient must be greater than zero")
	}
	if req.Fine < 0 {
		return ds.ShareholderInCalculation{}, fmt.Errorf("fine cannot be negative")
	}

	// --- ИСПРАВЛЕНИЕ ---
	// Используем .Updates() вместо .Save()
	updates := map[string]interface{}{
		"coefficient": req.Coefficient,
		"fine":        req.Fine,
	}

	err := r.db.Model(&ds.ShareholderInCalculation{}).
		Where("dividend_calculation_id = ? AND shareholder_id = ?", calculationID, shareholderID).
		Updates(updates).Error

	if err != nil {
		return ds.ShareholderInCalculation{}, err
	}
	
	// Теперь получаем обновленную запись, чтобы вернуть ее в ответе
	var updatedItem ds.ShareholderInCalculation
	err = r.db.Preload("Shareholder").
		Where("dividend_calculation_id = ? AND shareholder_id = ?", calculationID, shareholderID).
		First(&updatedItem).Error

	if err != nil {
		// Этого не должно произойти, так как мы только что ее обновили,
		// но на всякий случай проверяем
		return ds.ShareholderInCalculation{}, fmt.Errorf("%w: failed to retrieve updated item", ErrNotFound)
	}

	return updatedItem, nil
}