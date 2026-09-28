package service

import (
	"context"
	"errors"
	"fmt"

	"api/internal/config"
	"api/internal/model"
	"api/internal/repository"
	"api/common"
	"api/ent"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserAlreadyExists  = errors.New("user already exists")
)

type AuthService struct {
	userRepository repository.UserRepository
	passwordConfig *config.PasswordConfig
}

func NewAuthService(
	userRepository repository.UserRepository,
	passwordConfig *config.PasswordConfig,
) *AuthService {
	return &AuthService{
		userRepository: userRepository,
		passwordConfig: passwordConfig,
	}
}

func (s *AuthService) hashPassword(password string, cost int) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	return string(hashed), nil
}

func (s *AuthService) comparePassword(hashedPassword, password string) error {
	if err := bcrypt.CompareHashAndPassword(
		[]byte(hashedPassword),
		[]byte(password),
	); err != nil {
		return fmt.Errorf("compare hash and password: %w", err)
	}
	return nil
}

func (s *AuthService) Login(
	ctx context.Context,
	user *model.User,
) (*model.AuthResponse, error) {
	userEntity, err := s.userRepository.GetByEmail(ctx, user.Email)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("get user for login: %w", err)
	}

	if userEntity.State != common.ACTIVE {
		return nil, ErrInvalidCredentials
	}

	if err := s.comparePassword(
		userEntity.Password,
		user.Password,
	); err != nil {
		return nil, ErrInvalidCredentials
	}

	return &model.AuthResponse{
		ID:       userEntity.ID,
		Username: userEntity.Username,
		Email:    userEntity.Email,
	}, nil
}

func (s *AuthService) Register(
	ctx context.Context,
	user *model.User,
) (*model.AuthResponse, error) {
	_, err := s.userRepository.GetByEmail(ctx, user.Email)
	if err == nil {
		return nil, ErrUserAlreadyExists
	}
	if !ent.IsNotFound(err) {
		return nil, fmt.Errorf("check existing user: %w", err)
	}

	hashedPassword, err := s.hashPassword(
		user.Password,
		s.passwordConfig.Cost,
	)
	if err != nil {
		return nil, fmt.Errorf("hash register password: %w", err)
	}

	entity := &ent.User{
		Email:    user.Email,
		Username: user.Username,
		Password: hashedPassword,
		State:    common.ACTIVE,
	}

	createdUser, err := s.userRepository.Create(ctx, entity)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("create registered user: %w", err)
	}

	return &model.AuthResponse{
		ID:       createdUser.ID,
		Username: createdUser.Username,
		Email:    createdUser.Email,
	}, nil
}

func (s *AuthService) CurrentUser(
	ctx context.Context,
	id int,
) (*model.AuthResponse, error) {
	user, err := s.userRepository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get current user: %w", err)
	}
	return &model.AuthResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	}, nil
}
