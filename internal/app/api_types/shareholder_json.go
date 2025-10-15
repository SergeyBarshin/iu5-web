package api_types

import "shareholder-app/internal/app/ds"

// ShareholderRequest описывает JSON для создания/обновления акционера
type ShareholderRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Share       float64 `json:"share"`
}

// ShareholderResponse описывает JSON для одного акционера
type ShareholderResponse struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Share       float64 `json:"share"`
	ImageURL    *string `json:"image_url"` // Используем *string для nullable
}

// ConvertShareholderToResponse преобразует ds.Shareholder в ShareholderResponse
func ConvertShareholderToResponse(s ds.Shareholder) ShareholderResponse {
	var imageURL *string
	if s.ImageURL.Valid {
		imageURL = &s.ImageURL.String
	}

	return ShareholderResponse{
		ID:          s.ID,
		Name:        s.Name,
		Description: s.Description,
		Share:       s.Share,
		ImageURL:    imageURL,
	}
}

// ConvertShareholdersToResponse преобразует срез ds.Shareholder в срез ShareholderResponse
func ConvertShareholdersToResponse(shareholders []ds.Shareholder) []ShareholderResponse {
	responses := make([]ShareholderResponse, len(shareholders))
	for i, s := range shareholders {
		responses[i] = ConvertShareholderToResponse(s)
	}
	return responses
}