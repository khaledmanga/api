package services

import (
	"context"
	"errors"

	"api/src/config"
	"api/src/constants"
	"api/src/domain"
	"api/src/ent"
	"api/src/params/response"
	"api/src/repository"

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
		return "", err
	}

	return string(hashed), nil
}

func (s *AuthService) comparePassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword(
		[]byte(hashedPassword),
		[]byte(password),
	)
}

func (s *AuthService) Login(
	ctx context.Context,
	user *domain.User,
) (*response.AuthResponse, error) {
	userEntity, err := s.userRepository.GetByEmail(ctx, user.Email)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if userEntity.State != constants.ACTIVE {
		return nil, ErrInvalidCredentials
	}

	if err := s.comparePassword(
		userEntity.Password,
		user.Password,
	); err != nil {
		return nil, ErrInvalidCredentials
	}

	return &response.AuthResponse{
		ID:       userEntity.ID,
		Username: userEntity.Username,
		Email:    userEntity.Email,
	}, nil
}

func (s *AuthService) Register(
	ctx context.Context,
	user *domain.User,
) (*response.AuthResponse, error) {
	_, err := s.userRepository.GetByEmail(ctx, user.Email)
	if err == nil {
		return nil, ErrUserAlreadyExists
	}
	if !ent.IsNotFound(err) {
		return nil, err
	}

	hashedPassword, err := s.hashPassword(
		user.Password,
		s.passwordConfig.Cost,
	)
	if err != nil {
		return nil, err
	}

	entity := &ent.User{
		Email:    user.Email,
		Username: user.Username,
		Password: hashedPassword,
		State:    constants.ACTIVE,
	}

	createdUser, err := s.userRepository.Create(ctx, entity)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, ErrUserAlreadyExists
		}
		return nil, err
	}

	return &response.AuthResponse{
		ID:       createdUser.ID,
		Username: createdUser.Username,
		Email:    createdUser.Email,
	}, nil
}

func (s *AuthService) CurrentUser(
	ctx context.Context,
	id int,
) (*response.AuthResponse, error) {
	user, err := s.userRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &response.AuthResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	}, nil
}