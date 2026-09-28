package handler

import (
	"errors"
	"fmt"
	"strconv"

	"api/internal/repository"
	"github.com/gofiber/fiber/v3"
)

func currentUserID(c fiber.Ctx) (int, error) {
	userID, ok := c.Locals("userID").(int)
	if !ok || userID < 1 {
		return 0, fiber.ErrUnauthorized
	}
	return userID, nil
}

func positiveID(value string) (int, error) {
	id, err := strconv.Atoi(value)
	if err != nil || id < 1 {
		return 0, fiber.ErrBadRequest
	}
	return id, nil
}

func resourceError(err error) error {
	if errors.Is(err, repository.ErrNotFound) {
		return fiber.ErrNotFound
	}
	return fmt.Errorf("resource operation failed: %w", err)
}
