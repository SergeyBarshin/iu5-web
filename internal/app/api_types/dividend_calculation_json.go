// Файл: internal/app/api_types/dividend_calculation_json.go

package api_types

import (
	"shareholder-app/internal/app/ds"
	"time"
)

// --- Запросы ---

// CalculationUpdateRequest описывает JSON для обновления полей черновика
type CalculationUpdateRequest struct {
	TotalProfit float64 `json:"total_profit"`
}

// ModerationRequest описывает JSON для смены статуса модератором
type ModerationRequest struct {
	Status string `json:"status"` // "completed" или "rejected"
}


// --- Ответы ---

// CartInfoResponse описывает JSON для иконки "корзины"
type CartInfoResponse struct {
	DraftID uint  `json:"draft_id"`
	Count   int64 `json:"count"`
}

// CalculationItemResponse описывает одного акционера внутри ответа по расчету
type CalculationItemResponse struct {
	Shareholder ShareholderResponse `json:"shareholder"`
	Coefficient float64             `json:"coefficient"`
	Fine        float64             `json:"fine"`
	FinalDividend *float64           `json:"final_dividend,omitempty"` // omitempty скроет поле, если оно null
}

// CalculationResponse описывает краткий JSON для одного расчета (для списков)
type CalculationResponse struct {
	ID          int          `json:"id"`
	Status      string       `json:"status"`
	TotalProfit *float64     `json:"total_profit,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	SubmittedAt *time.Time   `json:"submitted_at,omitempty"`
	CompletedAt *time.Time   `json:"completed_at,omitempty"`
	Creator     UserResponse `json:"creator"`
	Moderator   *UserResponse `json:"moderator,omitempty"`
}

// CalculationDetailedResponse описывает полный JSON для одного расчета (для GET by ID)
type CalculationDetailedResponse struct {
	ID          int                     `json:"id"`
	Status      string                  `json:"status"`
	TotalProfit *float64                `json:"total_profit,omitempty"`
	CreatedAt   time.Time               `json:"created_at"`
	SubmittedAt *time.Time              `json:"submitted_at,omitempty"`
	CompletedAt *time.Time              `json:"completed_at,omitempty"`
	Creator     UserResponse            `json:"creator"`
	Moderator   *UserResponse           `json:"moderator,omitempty"`
	Items       []CalculationItemResponse `json:"items"` // Главное отличие - наличие Items
}


// --- Функции-конвертеры ---

// ConvertCalculationsToResponse конвертирует срез моделей БД в срез кратких JSON-ответов
func ConvertCalculationsToResponse(calculations []ds.DividendCalculation) []CalculationResponse {
	responses := make([]CalculationResponse, len(calculations))
	for i, c := range calculations {
		responses[i] = ConvertCalculationToResponse(c)
	}
	return responses
}

// ConvertCalculationToResponse конвертирует одну модель БД в краткий JSON-ответ
func ConvertCalculationToResponse(c ds.DividendCalculation) CalculationResponse {
	var totalProfit *float64
	if c.TotalProfit.Valid {
		totalProfit = &c.TotalProfit.Float64
	}

	var submittedAt *time.Time
	if c.SubmittedAt.Valid {
		submittedAt = &c.SubmittedAt.Time
	}

	var completedAt *time.Time
	if c.CompletedAt.Valid {
		completedAt = &c.CompletedAt.Time
	}

	var moderator *UserResponse
	if c.ModeratorID.Valid {
		modResp := ConvertUserToResponse(c.Moderator)
		moderator = &modResp
	}

	return CalculationResponse{
		ID:          c.ID,
		Status:      c.Status,
		TotalProfit: totalProfit,
		CreatedAt:   c.CreatedAt,
		SubmittedAt: submittedAt,
		CompletedAt: completedAt,
		Creator:     ConvertUserToResponse(c.Creator),
		Moderator:   moderator,
	}
}

// ConvertCalculationToDetailedResponse конвертирует расчет и его элементы в детальный JSON-ответ
func ConvertCalculationToDetailedResponse(c ds.DividendCalculation, items []ds.ShareholderInCalculation) CalculationDetailedResponse {
	// Используем базовый конвертер, чтобы не дублировать код
	baseResponse := ConvertCalculationToResponse(c)

	// Конвертируем items
	itemResponses := make([]CalculationItemResponse, len(items))
	for i, item := range items {
		var finalDividend *float64
		if item.FinalDividend.Valid {
			finalDividend = &item.FinalDividend.Float64
		}

		itemResponses[i] = CalculationItemResponse{
			Shareholder: ConvertShareholderToResponse(item.Shareholder),
			Coefficient:   item.Coefficient,
			Fine:          item.Fine,
			FinalDividend: finalDividend,
		}
	}

	return CalculationDetailedResponse{
		ID:          baseResponse.ID,
		Status:      baseResponse.Status,
		TotalProfit: baseResponse.TotalProfit,
		CreatedAt:   baseResponse.CreatedAt,
		SubmittedAt: baseResponse.SubmittedAt,
		CompletedAt: baseResponse.CompletedAt,
		Creator:     baseResponse.Creator,
		Moderator:   baseResponse.Moderator,
		Items:       itemResponses,
	}
}

type CalculationItemUpdateRequest struct {
	Coefficient float64 `json:"coefficient"`
	Fine        float64 `json:"fine"`
}