package repository

import (
	"fmt"
	"strings"
)

type Repository struct{}

func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}

type Shareholder struct {
	ID          int
	Name        string
	Description string
	Share       float64
	ImageURL    string
}

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

type Request struct {
	ID int
}

type RequestItem struct {
	RequestID     int
	ShareholderID int
	Coefficient   float64
	Fine          float64
}

var allRequests = []Request{
	{ID: 1},
}

var allRequestItems = []RequestItem{
	{RequestID: 1, ShareholderID: 1, Coefficient: 1.2, Fine: 1200},
	{RequestID: 1, ShareholderID: 2, Coefficient: 1.0, Fine: 2410},
}

func (r *Repository) GetRequest(id int) (Request, error) {
	for _, req := range allRequests {
		if req.ID == id {
			return req, nil
		}
	}
	return Request{}, fmt.Errorf("заявка с ID %d не найдена", id)
}

func (r *Repository) GetRequestItemsByRequestID(requestID int) ([]RequestItem, error) {
	var items []RequestItem
	for _, item := range allRequestItems {
		if item.RequestID == requestID {
			items = append(items, item)
		}
	}
	return items, nil
}

func (r *Repository) GetTotalRequestItemsCount() (int, error) {
	return len(allRequestItems), nil
}