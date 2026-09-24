package mapper

import (
	"api/src/dto"
	"api/src/ent"
	"api/src/request"
	"api/src/response"
)

type UserMapper struct{}

// For handler transfer request to pattern dto for service and repository
func (UserMapper) ToCreateDTO(req *request.CreateUserRequest) *dto.CreateUserDTO {
	return &dto.CreateUserDTO{
		Name:  req.Name,
		Email: req.Email,
	}
}

// For repository and service use to transfer record to specified pattern for handler
func (UserMapper) ToResponse(user *ent.User) *response.UserResponse {
	return &response.UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}
}
