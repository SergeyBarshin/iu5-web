package repository

import (
	"fmt"
	"strings"
)

type Repository struct{}

func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}

// Shareholder представляет модель акционера (аналог Material).
type Shareholder struct {
	ID          int
	Name        string
	Description string
	Share       float64
	ImageURL    string
}

// allShareholders содержит постоянный список всех акционеров.
var allShareholders = []Shareholder{
	{ID: 1, Name: "Рубчинский Георгий Александрович", Description: "Частный инвестор, поддерживает инициативы, направленные на минимизацию рисков.", Share: 11, ImageURL: "http://localhost:8080/static/img/rubchinskiy.png"}, // 8080/static = 9000
	{ID: 2, Name: "Дрёмин Иван Тимофеевич", Description: "Представитель инвестиционной группы, активно продвигает внедрение новых технологий.", Share: 33, ImageURL: "http://localhost:8080/static/img/dremin.png"},
	{ID: 3, Name: "Ляхов Григорий Алексеевич", Description: "Миноритарный акционер, заинтересован в сохранении корпоративных ценностей.", Share: 5, ImageURL: "http://localhost:8080/static/img/lyahov.png"},
	{ID: 4, Name: "Голубин Глеб Геннадьевич", Description: "Предприниматель с опытом в промышленности, делает ставку на эффективность и сокращение издержек.", Share: 10, ImageURL: "http://localhost:8080/static/img/golubin.png"},
}

func (r *Repository) GetShareholders() ([]Shareholder, error) {
	return allShareholders, nil
}

func (r *Repository) GetShareholder(id int) (Shareholder, error) {
	for _, s := range allShareholders {
		if s.ID == id {
			return s, nil
		}
	}
	return Shareholder{}, fmt.Errorf("акционер с ID %d не найден", id)
}

func (r *Repository) GetShareholdersByName(name string) ([]Shareholder, error) {
	trimmedQuery := strings.TrimSpace(strings.ToLower(name))
	if trimmedQuery == "" {
		return allShareholders, nil
	}
	var result []Shareholder
	for _, s := range allShareholders {
		if strings.Contains(strings.ToLower(s.Name), trimmedQuery) {
			result = append(result, s)
		}
	}
	return result, nil
}

// --- Логика для Заявки (Request) ---

// ИЗМЕНЕНО: RequestItem теперь содержит все данные, как OrderItem
type RequestItem struct {
	ShareholderID int
	Coefficient   float64 // Коэффициент корректировки доли
	Fine          float64 // Сумма штрафов (руб)
}

// allRequestItems - это статичный список элементов в заявке со всеми данными.
var allRequestItems = []RequestItem{
	{ShareholderID: 1, Coefficient: 1.2, Fine: 1200},
	{ShareholderID: 2, Coefficient: 1.0, Fine: 2410},
}

// GetRequestItems возвращает элементы заявки.
func (r *Repository) GetRequestItems() ([]RequestItem, error) {
	return allRequestItems, nil
}

// GetRequestItemsCount возвращает количество элементов в заявке.
func (r *Repository) GetRequestItemsCount() (int, error) {
	return len(allRequestItems), nil
}