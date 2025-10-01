package ds

// User представляет пользователя системы
type User struct {
	ID          uint   `gorm:"primaryKey" json:"id"`                               // уникальный идентификатор пользователя
	Login       string `gorm:"varchar(25);unique;not null" json:"login"`           // логин пользователя (уникальный, not null)
	Password    string `gorm:"varchar(100);not null" json:"-"`                     // пароль (не возвращается в JSON, not null)
	IsModerator bool   `gorm:"boolean;not null;default:false" json:"is_moderator"` // признак модератора (not null, по умолчанию false)
}
