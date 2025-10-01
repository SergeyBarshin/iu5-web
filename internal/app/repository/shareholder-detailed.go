package repository

import (
	"shareholder-app/internal/app/ds"
)

func (r *Repository) GetShareholderByID(id int) (ds.Shareholder, error) {
	var shareholder ds.Shareholder
	err := r.db.Where("id = ?", id).First(&shareholder).Error
	return shareholder, err
}