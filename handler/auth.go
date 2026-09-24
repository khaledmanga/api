package handler

import (
	"github.com/gofiber/fiber/v2"

	"api/src/mapper"
	"api/src/request"
)

type AuthHandler struct {
	mapper *mapper.UserMapper
}

func NewAuthHandler (mapper *mapper.UserMapper) *AuthHandler {
	return &AuthHandler {
		mapper: mapper
	}
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req request.CreateUserRequest

	if err != c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}

	data := h.mapper.ToCreateDTO(&req)

	//Service logic

	if err != nil {
		return err 
	}
}
