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

func (r *Repository) GetShareholdersWithFilter(nameFilter string) ([]ds.Shareholder, error) {
	var shareholders []ds.Shareholder
	query := r.db.Order("name ASC")
	if nameFilter != "" {
		query = query.Where("name ILIKE ?", "%"+nameFilter+"%")
	}
	err := query.Find(&shareholders).Error
	return shareholders, err
}

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

func (r *Repository) CreateShareholder(req api_types.ShareholderRequest) (ds.Shareholder, error) {
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

func (r *Repository) UpdateShareholder(id uint, req api_types.ShareholderRequest) (ds.Shareholder, error) {
	shareholder, err := r.GetShareholderByID(id)
	if err != nil {
		return ds.Shareholder{}, err
	}
	if req.Name == "" {
		return ds.Shareholder{}, errors.New("shareholder name cannot be empty")
	}
	if req.Share <= 0 || req.Share > 100 {
		return ds.Shareholder{}, errors.New("share must be between 0 and 100")
	}
	shareholder.Name = req.Name
	shareholder.Description = req.Description
	shareholder.Share = req.Share
	err = r.db.Save(&shareholder).Error
	return shareholder, err
}

func (r *Repository) DeleteShareholder(id uint) error {
	shareholder, err := r.GetShareholderByID(id)
	if err != nil {
		return err
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("shareholder_id = ?", id).Delete(&ds.ShareholderInCalculation{}).Error; err != nil {
			return err
		}
		if shareholder.ImageURL.Valid && shareholder.ImageURL.String != "" {
			objectName := r.mc.GetObjectNameFromURL(shareholder.ImageURL.String)
			_ = r.mc.DeleteImage(context.Background(), objectName)
		}
		if err := tx.Delete(&ds.Shareholder{}, id).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *Repository) UploadShareholderImage(id uint, file *multipart.FileHeader) (ds.Shareholder, error) {
	shareholder, err := r.GetShareholderByID(id)
	if err != nil {
		return ds.Shareholder{}, err
	}
	if shareholder.ImageURL.Valid && shareholder.ImageURL.String != "" {
		objectName := r.mc.GetObjectNameFromURL(shareholder.ImageURL.String)
		_ = r.mc.DeleteImage(context.Background(), objectName)
	}
	imageURL, err := r.mc.UploadImage(context.Background(), file, shareholder)
	if err != nil {
		return ds.Shareholder{}, fmt.Errorf("failed to upload image to minio: %w", err)
	}
	shareholder.ImageURL.String = imageURL
	shareholder.ImageURL.Valid = true
	err = r.db.Save(&shareholder).Error
	if err != nil {
		objectName := r.mc.GetObjectNameFromURL(imageURL)
		_ = r.mc.DeleteImage(context.Background(), objectName)
		return ds.Shareholder{}, fmt.Errorf("failed to save image URL to database: %w", err)
	}
	return shareholder, nil
}

func (r *Repository) AddShareholderToDraftCalculation(shareholderID, creatorID uint) error {
	if creatorID == 0 {
		return fmt.Errorf("%w: user is not authenticated", ErrNotAllowed)
	}
	draft, err := r.getOrCreateDraftCalculation(creatorID)
	if err != nil {
		return err
	}
	if _, err := r.GetShareholderByID(shareholderID); err != nil {
		return err
	}
	var existingLink ds.ShareholderInCalculation
	err = r.db.Where("dividend_calculation_id = ? AND shareholder_id = ?", draft.ID, shareholderID).First(&existingLink).Error
	if err == nil {
		return fmt.Errorf("%w: shareholder is already in the draft calculation", ErrAlreadyExists)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	newLink := ds.ShareholderInCalculation{
		DividendCalculationID: draft.ID,
		ShareholderID:         int(shareholderID),
		Coefficient:           1.0,
		Fine:                  0.0,
	}
	return r.db.Create(&newLink).Error
}

func (r *Repository) getOrCreateDraftCalculation(creatorID uint) (ds.DividendCalculation, error) {
	var calculation ds.DividendCalculation
	err := r.db.Where("creator_id = ? AND status = ?", creatorID, "draft").First(&calculation).Error
	if err == nil {
		return calculation, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
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
	return ds.DividendCalculation{}, err
}