package repository

import (
	"fmt"
	"shareholder-app/internal/app/api_types"
	"shareholder-app/internal/app/ds"
)

func (r *Repository) checkDraftAccess(calculationID, userID uint) error {
	var calculation ds.DividendCalculation
	err := r.db.Where("id = ?", calculationID).First(&calculation).Error
	if err != nil {
		return fmt.Errorf("%w: calculation not found", ErrNotFound)
	}
	if calculation.Status != "draft" {
		return fmt.Errorf("%w: can only modify a draft calculation", ErrNotAllowed)
	}
	if calculation.CreatorID != userID {
		return fmt.Errorf("%w: you are not the creator of this draft", ErrNotAllowed)
	}
	return nil
}

func (r *Repository) DeleteShareholderFromCalculation(calculationID, shareholderID, userID uint) error {
	if err := r.checkDraftAccess(calculationID, userID); err != nil {
		return err
	}
	result := r.db.Where("dividend_calculation_id = ? AND shareholder_id = ?", calculationID, shareholderID).
		Delete(&ds.ShareholderInCalculation{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: shareholder not found in this calculation draft", ErrNotFound)
	}
	return nil
}

func (r *Repository) UpdateShareholderInCalculation(calculationID, shareholderID, userID uint, req api_types.CalculationItemUpdateRequest) (ds.ShareholderInCalculation, error) {
	if err := r.checkDraftAccess(calculationID, userID); err != nil {
		return ds.ShareholderInCalculation{}, err
	}
	if req.Coefficient <= 0 {
		return ds.ShareholderInCalculation{}, fmt.Errorf("coefficient must be greater than zero")
	}
	if req.Fine < 0 {
		return ds.ShareholderInCalculation{}, fmt.Errorf("fine cannot be negative")
	}
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
	var updatedItem ds.ShareholderInCalculation
	err = r.db.Preload("Shareholder").
		Where("dividend_calculation_id = ? AND shareholder_id = ?", calculationID, shareholderID).
		First(&updatedItem).Error
	if err != nil {
		return ds.ShareholderInCalculation{}, fmt.Errorf("%w: failed to retrieve updated item", ErrNotFound)
	}
	return updatedItem, nil
}