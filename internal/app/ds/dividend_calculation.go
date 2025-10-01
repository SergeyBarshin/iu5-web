package ds

import (
	"database/sql"
	"time"
)

// DividendCalculation представляет заявку на расчет дивидендов
type DividendCalculation struct {
	ID          int          `gorm:"primaryKey"`                           // уникальный идентификатор расчета
	Status      string       `gorm:"type:varchar(50);not null"`            // статус: draft, deleted, submitted, completed, rejected
	TotalProfit sql.NullFloat64 `gorm:"type:numeric(15,2);default:null"`   // общая прибыль компании для расчета
	CreatedAt   time.Time    `gorm:"not null"`                             // дата создания
	SubmittedAt sql.NullTime `gorm:"default:null"`                         // дата формирования
	CompletedAt sql.NullTime `gorm:"default:null"`                         // дата завершения
	CreatorID   uint         `gorm:"not null"`                             // ID создателя
	ModeratorID sql.NullInt64  `gorm:"default:null"`                         // ID модератора (может быть null)
	Creator     User         `gorm:"foreignKey:CreatorID"`                 // связь с пользователем-создателем
	Moderator   User         `gorm:"foreignKey:ModeratorID"`               // связь с пользователем-модератором
}

func (DividendCalculation) TableName() string {
    return "dividend_calculations"
}