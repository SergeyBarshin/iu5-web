package ds

import "database/sql"

type Shareholder struct {
	ID          int            `gorm:"primaryKey"`                           // уникальный идентификатор акционера
	Name        string         `gorm:"type:varchar(100);not null"`           // ФИО акционера
	Description string         `gorm:"type:text"`                            // описание деятельности
	Share       float64        `gorm:"type:numeric(5,2);not null"`           // доля в компании в процентах
	ImageURL    sql.NullString `gorm:"type:varchar(255);default:null"`       // URL изображения (может быть null)
}

func (Shareholder) TableName() string {
    return "shareholders"
}