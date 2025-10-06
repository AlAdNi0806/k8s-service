package repository

import (
	"auth-service/internal/model"
	"context"

	"github.com/go-pg/pg/v10"
)

type UserRepository struct {
	db *pg.DB
}

func NewUserRepository(db *pg.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	_, err := r.db.Model(user).Insert()
	return err
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	user := &model.User{}
	err := r.db.Model(user).Where("email = ?", email).Select()
	if err != nil {
		return nil, err
	}
	return user, nil
}
