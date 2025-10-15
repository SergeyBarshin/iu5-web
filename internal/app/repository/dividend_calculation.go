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

func (r *Repository) GetCartInfo(userID uint) (uint, int64, error) {
	if userID == 0 {
		return 0, 0, nil
	}
	var calculation ds.DividendCalculation
	err := r.db.Where("creator_id = ? AND status = ?", userID, "draft").First(&calculation).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, 0, nil
		}
		return 0, 0, err
	}
	var count int64
	err = r.db.Model(&ds.ShareholderInCalculation{}).Where("dividend_calculation_id = ?", calculation.ID).Count(&count).Error
	if err != nil {
		return 0, 0, err
	}
	return uint(calculation.ID), count, nil
}

func (r *Repository) GetCalculationsList(from, to time.Time, status string, userID uint, isModerator bool) ([]ds.DividendCalculation, error) {
	var calculations []ds.DividendCalculation
	query := r.db.Preload("Creator").Preload("Moderator").
		Where("status NOT IN (?, ?)", "draft", "deleted")

	if !isModerator {
		query = query.Where("creator_id = ?", userID)
	}

	if !from.IsZero() {
		query = query.Where("submitted_at >= ?", from)
	}
	if !to.IsZero() {
		query = query.Where("submitted_at < ?", to.AddDate(0, 0, 1))
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Order("id DESC").Find(&calculations).Error
	return calculations, err
}

func (r *Repository) GetCalculationWithShareholders(id, userID uint, isModerator bool) (ds.DividendCalculation, []ds.ShareholderInCalculation, error) {
	var calculation ds.DividendCalculation
	err := r.db.Preload("Creator").Preload("Moderator").First(&calculation, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.DividendCalculation{}, nil, fmt.Errorf("%w: calculation with id %d", ErrNotFound, id)
		}
		return ds.DividendCalculation{}, nil, err
	}

	if calculation.Status == "deleted" {
		return ds.DividendCalculation{}, nil, fmt.Errorf("%w: calculation not accessible", ErrNotAllowed)
	}
	if !isModerator && calculation.CreatorID != userID {
		return ds.DividendCalculation{}, nil, fmt.Errorf("%w: you don't have access to this calculation", ErrNotAllowed)
	}

	var items []ds.ShareholderInCalculation
	err = r.db.Preload("Shareholder").Where("dividend_calculation_id = ?", id).Find(&items).Error
	if err != nil {
		return ds.DividendCalculation{}, nil, err
	}
	return calculation, items, nil
}

func (r *Repository) UpdateCalculation(id, userID uint, req api_types.CalculationUpdateRequest) (ds.DividendCalculation, error) {
	var calculation ds.DividendCalculation
	err := r.db.Where("id = ? AND status = 'draft'", id).First(&calculation).Error
	if err != nil {
		return ds.DividendCalculation{}, fmt.Errorf("%w: draft calculation not found", ErrNotFound)
	}
	if calculation.CreatorID != userID {
		return ds.DividendCalculation{}, fmt.Errorf("%w: you are not the creator of this draft", ErrNotAllowed)
	}
	calculation.TotalProfit = sql.NullFloat64{Float64: req.TotalProfit, Valid: true}
	if err := r.db.Save(&calculation).Error; err != nil {
		return ds.DividendCalculation{}, err
	}
	return calculation, nil
}

func (r *Repository) SubmitCalculation(id, userID uint) (ds.DividendCalculation, error) {
	var calculation ds.DividendCalculation
	err := r.db.First(&calculation, id).Error
	if err != nil {
		return ds.DividendCalculation{}, fmt.Errorf("%w: calculation not found", ErrNotFound)
	}
	if calculation.CreatorID != userID {
		return ds.DividendCalculation{}, fmt.Errorf("%w: only the creator can submit", ErrNotAllowed)
	}
	if calculation.Status != "draft" {
		return ds.DividendCalculation{}, fmt.Errorf("%w: can only submit a 'draft' calculation", ErrNotAllowed)
	}
	if !calculation.TotalProfit.Valid || calculation.TotalProfit.Float64 <= 0 {
		return ds.DividendCalculation{}, errors.New("total profit must be set and be greater than zero before submitting")
	}
	calculation.Status = "submitted"
	calculation.SubmittedAt = sql.NullTime{Time: time.Now(), Valid: true}
	if err := r.db.Save(&calculation).Error; err != nil {
		return ds.DividendCalculation{}, err
	}
	return calculation, nil
}

func (r *Repository) ModerateCalculation(id, userID uint, isModerator bool, req api_types.ModerationRequest) (ds.DividendCalculation, error) {
	if req.Status != "completed" && req.Status != "rejected" {
		return ds.DividendCalculation{}, errors.New("invalid status for moderation: must be 'completed' or 'rejected'")
	}
	if !isModerator {
		return ds.DividendCalculation{}, fmt.Errorf("%w: you are not a moderator", ErrNotAllowed)
	}
	var calculation ds.DividendCalculation
	err := r.db.First(&calculation, id).Error
	if err != nil {
		return ds.DividendCalculation{}, fmt.Errorf("%w: calculation not found", ErrNotFound)
	}
	if calculation.Status != "submitted" {
		return ds.DividendCalculation{}, fmt.Errorf("%w: can only moderate a 'submitted' calculation", ErrNotAllowed)
	}
	calculation.Status = req.Status
	calculation.ModeratorID = sql.NullInt64{Int64: int64(userID), Valid: true}
	calculation.CompletedAt = sql.NullTime{Time: time.Now(), Valid: true}
	if req.Status == "completed" {
		if err := r.calculateFinalDividends(&calculation); err != nil {
			return ds.DividendCalculation{}, fmt.Errorf("failed to calculate final dividends: %w", err)
		}
	}
	if err := r.db.Save(&calculation).Error; err != nil {
		return ds.DividendCalculation{}, err
	}
	return calculation, nil
}

func (r *Repository) calculateFinalDividends(calculation *ds.DividendCalculation) error {
	var items []ds.ShareholderInCalculation
	if err := r.db.Preload("Shareholder").Where("dividend_calculation_id = ?", calculation.ID).Find(&items).Error; err != nil {
		return err
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			dividend := (calculation.TotalProfit.Float64 * (item.Shareholder.Share / 100)) * item.Coefficient - item.Fine
			err := tx.Model(&ds.ShareholderInCalculation{}).
				Where("dividend_calculation_id = ? AND shareholder_id = ?", item.DividendCalculationID, item.ShareholderID).
				Update("final_dividend", dividend).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) DeleteCalculation(id, userID uint) error {
	var calculation ds.DividendCalculation
	err := r.db.First(&calculation, id).Error
	if err != nil {
		return fmt.Errorf("%w: calculation not found", ErrNotFound)
	}
	if calculation.CreatorID != userID || calculation.Status != "draft" {
		return fmt.Errorf("%w: only creator can delete their own draft", ErrNotAllowed)
	}
	result := r.db.Exec("UPDATE dividend_calculations SET status = ?, submitted_at = ? WHERE id = ?", "deleted", time.Now(), id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: calculation not found or already deleted", ErrNotFound)
	}
	return nil
}