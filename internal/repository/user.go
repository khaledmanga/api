package repository

import (
	"context"
	"fmt"

	"api/ent"
	"api/ent/user"
)

type UserRepository interface {
	Create(ctx context.Context, u *ent.User) (*ent.User, error)
	Update(ctx context.Context, u *ent.User) (*ent.User, error)
	Delete(ctx context.Context, id int) error
	GetByID(ctx context.Context, id int) (*ent.User, error)
	GetByEmail(ctx context.Context, email string) (*ent.User, error)
}

type userRepository struct {
	db *ent.Client
}

func NewUserRepository(db *ent.Client) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, u *ent.User) (*ent.User, error) {
	created, err := r.db.User.Create().
		SetEmail(u.Email).
		SetUsername(u.Username).
		SetPassword(u.Password).
		SetRole(u.Role).
		SetState(u.State).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return created, nil
}

func (r *userRepository) Update(ctx context.Context, u *ent.User) (*ent.User, error) {
	updated, err := r.db.User.UpdateOneID(u.ID).
		SetEmail(u.Email).
		SetUsername(u.Username).
		SetPassword(u.Password).
		SetRole(u.Role).
		SetState(u.State).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	return updated, nil
}

func (r *userRepository) Delete(ctx context.Context, id int) error {
	if err := r.db.User.DeleteOneID(id).Exec(ctx); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id int) (*ent.User, error) {
	u, err := r.db.User.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*ent.User, error) {
	u, err := r.db.User.Query().Where(user.Email(email)).Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return u, nil
}
