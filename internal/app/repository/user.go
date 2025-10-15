package repository

import (
	"errors"
	"fmt"
	"shareholder-app/internal/app/api_types"
	"shareholder-app/internal/app/ds"

	"gorm.io/gorm"
)

// GetUserByID находит пользователя по ID.
func (r *Repository) GetUserByID(id uint) (ds.User, error) {
	var user ds.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.User{}, fmt.Errorf("%w: user with id %d not found", ErrNotFound, id)
		}
		return ds.User{}, err
	}
	return user, nil
}

// GetUserByLogin находит пользователя по логину.
func (r *Repository) GetUserByLogin(login string) (ds.User, error) {
	var user ds.User
	err := r.db.Where("login = ?", login).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.User{}, fmt.Errorf("%w: user with login '%s' not found", ErrNotFound, login)
		}
		return ds.User{}, err
	}
	return user, nil
}

// RegisterUser создает нового пользователя. Пароль не хешируется, как в референсе.
func (r *Repository) RegisterUser(req api_types.UserRegisterRequest) (ds.User, error) {
	// Проверка на существование пользователя
	_, err := r.GetUserByLogin(req.Login)
	if !errors.Is(err, ErrNotFound) { // Если ошибка НЕ "не найдено", значит, пользователь существует или другая проблема
		if err == nil {
			return ds.User{}, fmt.Errorf("%w: user with login '%s' already exists", ErrAlreadyExists, req.Login)
		}
		return ds.User{}, err
	}

	// Создаем нового пользователя
	newUser := ds.User{
		Login:       req.Login,
		Password:    req.Password, // Сохраняем пароль в открытом виде
		IsModerator: false,        // Новые пользователи по умолчанию не модераторы
	}

	if err := r.db.Create(&newUser).Error; err != nil {
		return ds.User{}, fmt.Errorf("failed to create user: %w", err)
	}

	return newUser, nil
}

// LoginUser (SignIn в референсе) аутентифицирует пользователя.
func (r *Repository) LoginUser(req api_types.UserLoginRequest) (ds.User, error) {
	user, err := r.GetUserByLogin(req.Login)
	if err != nil {
		return ds.User{}, errors.New("invalid login or password")
	}

	// Сравниваем пароли напрямую
	if user.Password != req.Password {
		return ds.User{}, errors.New("invalid login or password")
	}

	// "Авторизуем" пользователя, сохраняя его ID в репозитории
	r.SetUserID(user.ID)
	return user, nil
}

// UpdateProfile (ChangeProfile в референсе) обновляет данные пользователя.
func (r *Repository) UpdateProfile(id uint, req api_types.UserUpdateRequest) (ds.User, error) {
	user, err := r.GetUserByID(id)
	if err != nil {
		return ds.User{}, err
	}
	
	// Обновляем только разрешенные поля. Логин и роль менять не даем.
	// В референсе можно было менять пароль, добавим это.
	if req.Password != "" {
		user.Password = req.Password
	}
	// Добавим возможность менять ФИО, если бы оно было
	// user.FullName = req.FullName 

	if err := r.db.Save(&user).Error; err != nil {
		return ds.User{}, err
	}
	return user, nil
}