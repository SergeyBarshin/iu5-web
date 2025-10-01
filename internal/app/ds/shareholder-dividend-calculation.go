package ds

import "database/sql"

// ShareholderInCalculation представляет связь "акционер ↔ расчет" (многие ко многим)
type ShareholderInCalculation struct {
	DividendCalculationID int `gorm:"not null;uniqueIndex:idx_calculation_shareholder"`
	ShareholderID         int `gorm:"not null;uniqueIndex:idx_calculation_shareholder"`

	// Дополнительные поля для этой связи
	Coefficient   float64         `gorm:"type:numeric(5,2);not null"`           // коэффициент корректировки
	Fine          float64         `gorm:"type:numeric(15,2);not null"`          // сумма штрафов
	FinalDividend sql.NullFloat64 `gorm:"type:numeric(15,2);default:null"`      // итоговые дивиденды (рассчитывается при завершении)

	// Связи с моделями
	DividendCalculation DividendCalculation `gorm:"foreignKey:DividendCalculationID"`
	Shareholder         Shareholder         `gorm:"foreignKey:ShareholderID"`
}

func (ShareholderInCalculation) TableName() string {
    return "shareholder_dividend_calculations"
}