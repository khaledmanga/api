package handler

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"api/internal/cache"
	"api/internal/model"
	"api/internal/repository"
	"api/internal/service"
	"api/ent"

	"github.com/gofiber/fiber/v3"
)

const sessionLifetime = 7 * 24 * time.Hour

type AuthHandler struct {
	authService           *service.AuthService
	cacheClient           *cache.Client
	sessionCookieSameSite string
}

func NewAuthHandler(
	authService *service.AuthService,
	cacheClient *cache.Client,
	sessionCookieSameSite string,
) *AuthHandler {
	return &AuthHandler{
		authService:           authService,
		cacheClient:           cacheClient,
		sessionCookieSameSite: sessionCookieSameSite,
	}
}

func generateSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func setCookie(c fiber.Ctx, name, value string, maxAge int, sameSite string) {
	c.Cookie(&fiber.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HTTPOnly: true,
		Secure:   true,
		SameSite: sameSite,
	})
}

func (h *AuthHandler) createSession(c fiber.Ctx, userID int) error {
	token, err := generateSessionToken()
	if err != nil {
		return fmt.Errorf("generate session token: %w", err)
	}
	if err := h.cacheClient.Set(c.Context(), token, userID, sessionLifetime); err != nil {
		return fmt.Errorf("persist session: %w", err)
	}
	setCookie(
		c,
		"session",
		token,
		int(sessionLifetime.Seconds()),
		h.sessionCookieSameSite,
	)
	return nil
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	var params model.LoginRequest
	if err := c.Bind().Body(&params); err != nil {
		return fiber.ErrBadRequest
	}
	params.Email = strings.TrimSpace(params.Email)
	if params.Email == "" || len(params.Email) > 254 ||
		params.Password == "" || len(params.Password) > 72 {
		return fiber.ErrBadRequest
	}

	userModel := &model.User{
		Email:    params.Email,
		Password: params.Password,
	}

	result, err := h.authService.Login(c.Context(), userModel)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return fiber.ErrUnauthorized
		}
		return err
	}

	if err := h.createSession(c, result.ID); err != nil {
		return err
	}

	return c.JSON(result)
}

func (h *AuthHandler) Register(c fiber.Ctx) error {
	var params model.CreateUserRequest
	if err := c.Bind().Body(&params); err != nil {
		return fiber.ErrBadRequest
	}
	params.Username = strings.TrimSpace(params.Username)
	params.Email = strings.TrimSpace(params.Email)
	if len(params.Username) == 0 || utf8.RuneCountInString(params.Username) > 80 ||
		len(params.Email) == 0 || len(params.Email) > 254 ||
		len(params.Password) < 8 || len(params.Password) > 72 {
		return fiber.ErrBadRequest
	}

	userModel := &model.User{
		Username: params.Username,
		Email:    params.Email,
		Password: params.Password,
	}

	result, err := h.authService.Register(c.Context(), userModel)
	if err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			return fiber.ErrConflict
		}
		return err
	}

	if err := h.createSession(c, result.ID); err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

func (h *AuthHandler) Me(c fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	user, err := h.authService.CurrentUser(c.Context(), userID)
	if err != nil {
		if ent.IsNotFound(err) || errors.Is(err, repository.ErrNotFound) {
			return fiber.ErrUnauthorized
		}
		return fmt.Errorf("load current user: %w", err)
	}
	return c.JSON(user)
}

func (h *AuthHandler) Logout(c fiber.Ctx) error {
	token := c.Cookies("session")
	if token != "" {
		if err := h.cacheClient.Del(c.Context(), token); err != nil {
			return fiber.ErrInternalServerError
		}
	}
	c.Cookie(&fiber.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HTTPOnly: true,
		Secure:   true,
		SameSite: h.sessionCookieSameSite,
	})
	return c.SendStatus(fiber.StatusNoContent)
}
