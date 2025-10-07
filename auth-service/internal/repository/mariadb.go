// internal/repository/mariadb.go
package repository

import (
	"database/sql"
	"fmt"

	"auth-service/internal/model"

	_ "github.com/go-sql-driver/mysql"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(email, password string) error {
	_, err := r.db.Exec("INSERT INTO users (email, password) VALUES (?, ?)", email, password)
	return err
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRow("SELECT id, email, password FROM users WHERE email = ?", email).
		Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}
	return user, nil
}
