// Файл: internal/app/repository/user.go

package repository

import (
	"errors"
	"fmt"
	"os"
	"shareholder-app/internal/app/api_types"
	"shareholder-app/internal/app/ds"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

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

func (r *Repository) RegisterUser(req api_types.UserRegisterRequest) (ds.User, error) {
	_, err := r.GetUserByLogin(req.Login)
	if !errors.Is(err, ErrNotFound) {
		if err == nil {
			return ds.User{}, fmt.Errorf("%w: user with login '%s' already exists", ErrAlreadyExists, req.Login)
		}
		return ds.User{}, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return ds.User{}, fmt.Errorf("failed to hash password: %w", err)
	}

	newUser := ds.User{
		Login:       req.Login,
		Password:    string(hashedPassword),
		IsModerator: false,
	}

	if err := r.db.Create(&newUser).Error; err != nil {
		return ds.User{}, fmt.Errorf("failed to create user: %w", err)
	}
	return newUser, nil
}

func (r *Repository) SignIn(req api_types.UserLoginRequest) (string, error) {
	user, err := r.GetUserByLogin(req.Login)
	if err != nil {
		return "", errors.New("invalid login or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return "", errors.New("invalid login or password")
	}

	token, err := generateToken(user)
	if err != nil {
		return "", err
	}
	return token, nil
}

func generateToken(user ds.User) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["user_id"] = user.ID
	claims["is_moderator"] = user.IsModerator
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_KEY")))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return tokenString, nil
}

func (r *Repository) UpdateProfile(id uint, req api_types.UserUpdateRequest) (ds.User, error) {
	user, err := r.GetUserByID(id)
	if err != nil {
		return ds.User{}, err
	}
	
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return ds.User{}, fmt.Errorf("failed to hash password: %w", err)
		}
		user.Password = string(hashedPassword)
	}

	if err := r.db.Save(&user).Error; err != nil {
		return ds.User{}, err
	}
	return user, nil
}