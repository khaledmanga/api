package middleware

import (
	"errors"
	"fmt"
	"strconv"

	"api/src/config"
	"api/src/constants"
	"api/src/ent"
	"api/src/repository"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

func Authentication(redisClient *config.RedisClient, users repository.UserRepository) fiber.Handler {
	return func(c fiber.Ctx) error {
		token := c.Cookies("session")
		if token == "" {
			return fiber.ErrUnauthorized
		}
		userID, err := redisClient.Get(c.Context(), token)
		if errors.Is(err, redis.Nil) {
			return fiber.ErrUnauthorized
		}
		if err != nil {
			return fmt.Errorf("read session from Redis: %w", err)
		}
		id, err := strconv.Atoi(userID)
		if err != nil || id < 1 {
			return fiber.ErrUnauthorized
		}
		user, err := users.GetByID(c.Context(), id)
		if ent.IsNotFound(err) || err == nil && user.State != constants.ACTIVE {
			if deleteErr := redisClient.Del(c.Context(), token); deleteErr != nil {
				return fmt.Errorf("revoke invalid session: %w", deleteErr)
			}
			return fiber.ErrUnauthorized
		}
		if err != nil {
			return fmt.Errorf("load session user: %w", err)
		}
		c.Locals("userID", id)
		return c.Next()
	}
}
