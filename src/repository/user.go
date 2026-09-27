package repository

import (
	"context"

	"api/src/ent"
	"api/src/ent/user"
)

type UserRepository interface {
	Create(ctx context.Context, u *ent.User) (*ent.User, error)
	Update(ctx context.Context, u *ent.User) (*ent.User, error)
	Delete(ctx context.Context, id int) error
	GetByID(ctx context.Context, id int) (*ent.User, error)
	GetByEmail(ctx context.Context, email string) (*ent.User, error)
}

type userRepository struct{ db *ent.Client }

func NewUserRepository(db *ent.Client) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, u *ent.User) (*ent.User, error) {
	return r.db.User.Create().SetEmail(u.Email).SetUsername(u.Username).SetPassword(u.Password).SetRole(u.Role).SetState(u.State).Save(ctx)
}

func (r *userRepository) Update(ctx context.Context, u *ent.User) (*ent.User, error) {
	return r.db.User.UpdateOneID(u.ID).SetEmail(u.Email).SetUsername(u.Username).SetPassword(u.Password).SetRole(u.Role).SetState(u.State).Save(ctx)
}

func (r *userRepository) Delete(ctx context.Context, id int) error {
	return r.db.User.DeleteOneID(id).Exec(ctx)
}

func (r *userRepository) GetByID(ctx context.Context, id int) (*ent.User, error) {
	return r.db.User.Get(ctx, id)
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*ent.User, error) {
	return r.db.User.Query().Where(user.EmailEQ(email)).Only(ctx)
}
