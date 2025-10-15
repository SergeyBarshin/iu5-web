package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"shareholder-app/internal/app/api_types"
	"shareholder-app/internal/app/ds"
	"time"

	"gorm.io/gorm"
)

// GetCartInfo возвращает ID черновика и количество акционеров в нем.
// Аналог GetResearchCount и части GetResearchCart из референса.
func (r *Repository) GetCartInfo() (uint, int64, error) {
	creatorID := r.GetUserID()
	if creatorID == 0 {
		return 0, 0, nil // Если пользователь не "авторизован", корзина пуста
	}

	// Ищем существующий черновик
	var calculation ds.DividendCalculation
	err := r.db.Where("creator_id = ? AND status = ?", creatorID, "draft").First(&calculation).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, 0, nil // Черновика нет, это нормальная ситуация
		}
		return 0, 0, err
	}

	// Считаем количество акционеров в найденном черновике
	var count int64
	err = r.db.Model(&ds.ShareholderInCalculation{}).Where("dividend_calculation_id = ?", calculation.ID).Count(&count).Error
	if err != nil {
		return 0, 0, err
	}

	return uint(calculation.ID), count, nil
}

// GetCalculationsList возвращает список расчетов с фильтрацией.
// Не возвращает "draft" и "deleted".
// Аналог GetResearches из референса.
func (r *Repository) GetCalculationsList(from, to time.Time, status string) ([]ds.DividendCalculation, error) {
	var calculations []ds.DividendCalculation

	// Загружаем связанные данные Creator и Moderator сразу, чтобы избежать N+1 запросов
	query := r.db.Preload("Creator").Preload("Moderator").
		Where("status NOT IN (?, ?)", "draft", "deleted")

	if !from.IsZero() {
		// submitted_at - это дата формирования, по ней и фильтруем
		query = query.Where("submitted_at >= ?", from)
	}
	if !to.IsZero() {
		// Добавляем 24 часа, чтобы включить весь день "to"
		query = query.Where("submitted_at < ?", to.AddDate(0, 0, 1))
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Order("id DESC").Find(&calculations).Error
	return calculations, err
}

// GetCalculationWithShareholders получает один расчет со всеми его акционерами.
// Аналог GetResearchPlanets из референса.
func (r *Repository) GetCalculationWithShareholders(id uint) (ds.DividendCalculation, []ds.ShareholderInCalculation, error) {
	var calculation ds.DividendCalculation
	err := r.db.Preload("Creator").Preload("Moderator").First(&calculation, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.DividendCalculation{}, nil, fmt.Errorf("%w: calculation with id %d", ErrNotFound, id)
		}
		return ds.DividendCalculation{}, nil, err
	}

	// Проверка, что пользователь не пытается получить удаленный или чужой черновик
	// (логику доступа расширим в следующих ЛР, пока просто базовая проверка)
	if calculation.Status == "deleted" || (calculation.Status == "draft" && calculation.CreatorID != r.GetUserID()) {
		return ds.DividendCalculation{}, nil, fmt.Errorf("%w: calculation not accessible", ErrNotAllowed)
	}

	// Загружаем все связанные записи из М-М таблицы, а также самих акционеров
	var items []ds.ShareholderInCalculation
	err = r.db.Preload("Shareholder").Where("dividend_calculation_id = ?", id).Find(&items).Error
	if err != nil {
		return ds.DividendCalculation{}, nil, err
	}

	return calculation, items, nil
}

// UpdateCalculation обновляет поля расчета (только для черновика).
// Аналог ChangeResearch из референса.
func (r *Repository) UpdateCalculation(id uint, req api_types.CalculationUpdateRequest) (ds.DividendCalculation, error) {
	var calculation ds.DividendCalculation
	err := r.db.Where("id = ? AND status = 'draft'", id).First(&calculation).Error
	if err != nil {
		return ds.DividendCalculation{}, fmt.Errorf("%w: draft calculation not found", ErrNotFound)
	}

	// Проверяем, что текущий пользователь является создателем этого черновика
	if calculation.CreatorID != r.GetUserID() {
		return ds.DividendCalculation{}, fmt.Errorf("%w: you are not the creator of this draft", ErrNotAllowed)
	}

	// Обновляем только разрешенные поля
	calculation.TotalProfit = sql.NullFloat64{Float64: req.TotalProfit, Valid: true}

	if err := r.db.Save(&calculation).Error; err != nil {
		return ds.DividendCalculation{}, err
	}
	return calculation, nil
}

// SubmitCalculation меняет статус с 'draft' на 'submitted'.
// Аналог FormResearch из референса.
func (r *Repository) SubmitCalculation(id uint) (ds.DividendCalculation, error) {
	var calculation ds.DividendCalculation
	err := r.db.First(&calculation, id).Error
	if err != nil {
		return ds.DividendCalculation{}, fmt.Errorf("%w: calculation not found", ErrNotFound)
	}

	// Проверяем права и статус
	if calculation.CreatorID != r.GetUserID() {
		return ds.DividendCalculation{}, fmt.Errorf("%w: only the creator can submit", ErrNotAllowed)
	}
	if calculation.Status != "draft" {
		return ds.DividendCalculation{}, fmt.Errorf("%w: can only submit a 'draft' calculation", ErrNotAllowed)
	}
	// Проверка на обязательные поля перед формированием
	if !calculation.TotalProfit.Valid || calculation.TotalProfit.Float64 <= 0 {
		return ds.DividendCalculation{}, errors.New("total profit must be set and be greater than zero before submitting")
	}

	// Обновляем статус и дату формирования
	calculation.Status = "submitted"
	calculation.SubmittedAt = sql.NullTime{Time: time.Now(), Valid: true}

	if err := r.db.Save(&calculation).Error; err != nil {
		return ds.DividendCalculation{}, err
	}
	return calculation, nil
}

// ModerateCalculation меняет статус с 'submitted' на 'completed' или 'rejected'.
// Аналог ModerateResearch из референса.
func (r *Repository) ModerateCalculation(id uint, req api_types.ModerationRequest) (ds.DividendCalculation, error) {
	if req.Status != "completed" && req.Status != "rejected" {
		return ds.DividendCalculation{}, errors.New("invalid status for moderation: must be 'completed' or 'rejected'")
	}

	// В этой ЛР модератор у нас захардкожен, но логику проверки прав оставим
	// const moderatorID = 2 // Предположим, модератор имеет ID=2
	// if r.GetUserID() != moderatorID {
	// 	return ds.DividendCalculation{}, fmt.Errorf("%w: you are not a moderator", ErrNotAllowed)
	// }

	var calculation ds.DividendCalculation
	err := r.db.First(&calculation, id).Error
	if err != nil {
		return ds.DividendCalculation{}, fmt.Errorf("%w: calculation not found", ErrNotFound)
	}

	if calculation.Status != "submitted" {
		return ds.DividendCalculation{}, fmt.Errorf("%w: can only moderate a 'submitted' calculation", ErrNotAllowed)
	}

	// Обновляем статус, модератора и дату завершения
	// moderatorID будет тем же, что и GetUserID(), т.к. создатель и модератор у нас один
	calculation.Status = req.Status
	calculation.ModeratorID = sql.NullInt64{Int64: int64(r.GetUserID()), Valid: true}
	calculation.CompletedAt = sql.NullTime{Time: time.Now(), Valid: true}

	// Если заявка ЗАВЕРШЕНА, запускаем расчет дивидендов
	if req.Status == "completed" {
		if err := r.calculateFinalDividends(&calculation); err != nil {
			return ds.DividendCalculation{}, fmt.Errorf("failed to calculate final dividends: %w", err)
		}
	}
	
	// Сохраняем все изменения одним запросом
	if err := r.db.Save(&calculation).Error; err != nil {
		return ds.DividendCalculation{}, err
	}
	
	return calculation, nil
}

// calculateFinalDividends - та самая формула из ЛР2.
// Вызывается внутри транзакции при завершении расчета.
func (r *Repository) calculateFinalDividends(calculation *ds.DividendCalculation) error {
	var items []ds.ShareholderInCalculation
	// Загружаем связанные записи M-M
	if err := r.db.Preload("Shareholder").Where("dividend_calculation_id = ?", calculation.ID).Find(&items).Error; err != nil {
		return err
	}

	// Обернем все обновления в транзакцию для надежности
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			// Формула: (Общая прибыль * (Доля / 100)) * Коэффициент - Штрафы
			dividend := (calculation.TotalProfit.Float64 * (item.Shareholder.Share / 100)) * item.Coefficient - item.Fine

			err := tx.Model(&ds.ShareholderInCalculation{}).
				Where("dividend_calculation_id = ? AND shareholder_id = ?", item.DividendCalculationID, item.ShareholderID).
				Update("final_dividend", dividend).Error

			if err != nil {
				return err // Если ошибка, транзакция будет отменена
			}
		}
		return nil // Если все успешно, транзакция будет закоммичена
	})
}

// DeleteCalculation выполняет логическое удаление расчета.
// Для этого используется Raw SQL, как того требует задание.
// Аналог DeleteCalculation из референса, но с логикой soft-delete.
func (r *Repository) DeleteCalculation(id uint) error {
	var calculation ds.DividendCalculation
	// Сначала получаем расчет, чтобы проверить права
	err := r.db.First(&calculation, id).Error
	if err != nil {
		return fmt.Errorf("%w: calculation not found", ErrNotFound)
	}

	// Удалять можно только свой черновик
	if calculation.CreatorID != r.GetUserID() || calculation.Status != "draft" {
		return fmt.Errorf("%w: only creator can delete their own draft", ErrNotAllowed)
	}
	
	// Используем Raw SQL для обновления статуса
	// GORM безопасно обрабатывает параметры для предотвращения SQL-инъекций
	result := r.db.Exec("UPDATE dividend_calculations SET status = ?, submitted_at = ? WHERE id = ?", "deleted", time.Now(), id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: calculation not found or already deleted", ErrNotFound)
	}
	
	return nil
}