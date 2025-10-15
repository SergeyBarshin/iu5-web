package api_types

import "shareholder-app/internal/app/ds"

type UserRegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserLoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID          uint   `json:"id"`
	Login       string `json:"login"`
	IsModerator bool   `json:"is_moderator"`
}

func ConvertUserToResponse(user ds.User) UserResponse {
	return UserResponse{
		ID:          user.ID,
		Login:       user.Login,
		IsModerator: user.IsModerator,
	}
}
type UserUpdateRequest struct {
	Password string `json:"password"`
	// Здесь могли бы быть и другие поля, например, FullName
}