package mapper

import (
	"api/src/domain"
	"api/src/ent"
	"api/src/params/request"
	"api/src/params/response"
)

type UserMapper struct{}

func NewUserMapper() *UserMapper {
	return &UserMapper{}
}

func (um *UserMapper) DomainToEnt(user *domain.User) *ent.User {
	return &ent.User{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Password: user.Password,
	}
}

func (um *UserMapper) EntToDomain(user *ent.User) *domain.User {
	return &domain.User{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Password: user.Password,
	}
}

func (um *UserMapper) CreateUserParamsToDomain(params *request.CreateUserParams) *domain.User {
	return &domain.User{
		Username: params.Username,
		Email:    params.Email,
		Password: params.Password,
	}
}

func (um *UserMapper) DomainToAuthResponse(user *domain.User, role string) *response.AuthResponse {
	return &response.AuthResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	}
}
func (um *UserMapper) LoginParamsToDomain(params *request.LoginParams) *domain.User {
	return &domain.User{
		Email:    params.Email,
		Password: params.Password,
	}
}
